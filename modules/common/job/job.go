/*
Copyright 2021 Red Hat

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package job

import (
	"context"
	"fmt"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	"github.com/openstack-k8s-operators/lib-common/modules/common/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"errors"
)

const hashAnnotation = "hash"

// NewJob returns an initialized Job.
func NewJob(
	job *batchv1.Job,
	jobType string,
	preserve bool,
	timeout int,
	beforeHash string,
) *Job {

	return &Job{
		job:        job,
		jobType:    jobType,
		preserve:   preserve,
		timeout:    time.Duration(timeout) * time.Second, // timeout to set in s to reconcile
		beforeHash: beforeHash,
		changed:    false,
	}
}

// createJob - creates job, reconciles after Xs if object won't exist.
func (j *Job) createJob(
	ctx context.Context,
	h *helper.Helper,
) (ctrl.Result, error) {
	op, err := controllerutil.CreateOrPatch(ctx, h.GetClient(), j.job, func() error {
		err := controllerutil.SetControllerReference(h.GetBeforeObject(), j.job, h.GetScheme())
		if err != nil {
			return err
		}
		// Add the job hash as an annotation, this is used by the DeleteAllSucceededJobs
		// to filter on jobs by hash
		j.job.Annotations = util.MergeStringMaps(j.job.Labels, map[string]string{hashAnnotation: j.hash})

		return nil
	})
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			h.GetLogger().Info(fmt.Sprintf("Job %s not found, reconcile in %s", j.job.Name, j.timeout))
			return ctrl.Result{RequeueAfter: j.timeout}, nil
		}
		return ctrl.Result{}, err
	}
	if op != controllerutil.OperationResultNone {
		h.GetLogger().Info(fmt.Sprintf("Job %s %s - %s", j.jobType, j.job.Name, op))
		return ctrl.Result{RequeueAfter: j.timeout}, nil
	}

	return ctrl.Result{}, nil
}

//
// DoJob - run a job if the hashBefore and hash is different. If there is an existing job, wait for the job
// to finish. Right now we do not expect the job to change while running.
//
func (j *Job) DoJob(
	ctx context.Context,
	h *helper.Helper,
) (ctrl.Result, error) {
	var ctrlResult ctrl.Result
	var err error

	j.hash, err = util.ObjectHash(j.job)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error calculating %s hash: %v", j.jobType, err)
	}

	// if the hash changed the job should run
	if j.beforeHash != j.hash {
		j.changed = true
	}

	//
	// Check if this job already exists
	//
	_, err = GetJobWithName(ctx, h, j.job.Name, j.job.Namespace)
	if err != nil && !k8s_errors.IsNotFound(err) {
		return ctrl.Result{}, err
	}

	wait := false
	if !k8s_errors.IsNotFound(err) {
		// if job exist, wait for it to finish
		// for now we do not expect the job to change while running
		wait = true
	} else if j.changed {
		// if job changed, create it and wait for it to finish
		ctrlResult, err = j.createJob(ctx, h)
		if err != nil {
			return ctrlResult, err
		}
		wait = true
	}

	if wait {
		ctrlResult, err := waitOnJob(ctx, h, j.job.Name, j.job.Namespace, j.timeout)
		if (ctrlResult != ctrl.Result{}) {
			return ctrlResult, nil
		}
		if err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

// HasChanged func
func (j *Job) HasChanged() bool {
	return j.changed
}

// GetHash func
func (j *Job) GetHash() string {
	return j.hash
}

// SetTimeout defines the duration used for requeueing while waiting for the job
// to finish.
func (j *Job) SetTimeout(timeout time.Duration) {
	j.timeout = timeout
}

// DeleteJob func
// kclient required to properly cleanup the job depending pods with DeleteOptions
func DeleteJob(
	ctx context.Context,
	h *helper.Helper,
	name string,
	namespace string,
) error {
	foundJob, err := h.GetKClient().BatchV1().Jobs(namespace).Get(ctx, name, metav1.GetOptions{})
	if err == nil {
		err := deleteJobByObject(ctx, h, *foundJob)
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}

func deleteJobByObject(ctx context.Context, h *helper.Helper, job batchv1.Job) error {
	util.LogForObject(h, "Deleting Job", h.GetBeforeObject(), "Job.Namespace", job.Namespace, "Job.Name", job.Name)
	background := metav1.DeletePropagationBackground
	err := h.GetKClient().BatchV1().Jobs(job.Namespace).Delete(
		ctx, job.Name, metav1.DeleteOptions{PropagationPolicy: &background})
	if err != nil {
		return err
	}
	return nil
}

// DeleteAllSucceededJobs deletes all the jobs that matching the criterias:
// 1. owned by the caller
// 2. the Job's hash is in the jobHashes list
// 3. the Job is succeeded i.e. job.Status.Succeeded > 0
func DeleteAllSucceededJobs(
	ctx context.Context,
	h *helper.Helper,
	jobHashes []string,
) error {
	jobs := &batchv1.JobList{}

	err := h.GetClient().List(ctx, jobs, client.InNamespace(h.GetBefore().GetNamespace()))
	if err != nil {
		return err
	}
	h.GetBeforeObject()

	for _, job := range jobs.Items {
		// 1. check if caller owns it
		controller := metav1.GetControllerOf(&job)
		if controller == nil {
			continue
		}
		kind := h.GetBeforeObject().GetObjectKind().GroupVersionKind().Kind
		name := h.GetBeforeObject().GetName()
		if controller.Kind != kind || controller.Name != name {
			continue
		}
		// 2. check if Job's hash is in jobHashes
		v, ok := job.Annotations[hashAnnotation]
		if !ok {
			continue
		}
		if !util.StringInSlice(v, jobHashes) {
			continue
		}
		// 3. check if succeeded
		if job.Status.Succeeded <= 0 {
			continue
		}

		err = deleteJobByObject(ctx, h, job)
		if err != nil {
			return err
		}
	}
	return nil
}

// waitOnJob func -  returns true if the job
func waitOnJob(
	ctx context.Context,
	h *helper.Helper,
	name string,
	namespace string,
	timeout time.Duration,
) (ctrl.Result, error) {
	// Check if this Job already exists
	job, err := GetJobWithName(ctx, h, name, namespace)
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			h.GetLogger().Info("Job was not found.")
			return ctrl.Result{RequeueAfter: timeout}, nil
		}
		h.GetLogger().Info("WaitOnJob err")
		return ctrl.Result{}, err
	}

	if job.Status.Active > 0 {
		h.GetLogger().Info("Job Status Active... requeuing")
		return ctrl.Result{RequeueAfter: timeout}, nil
	} else if job.Status.Succeeded > 0 {
		h.GetLogger().Info("Job Status Successful")
		return ctrl.Result{}, nil
	} else if job.Status.Failed > 0 {
		h.GetLogger().Info("Job Status Failed")
		return ctrl.Result{}, k8s_errors.NewInternalError(errors.New("Job Failed. Check job logs"))
	}
	h.GetLogger().Info("Job Status incomplete... requeuing")
	return ctrl.Result{RequeueAfter: timeout}, nil
}

// GetJobWithName func
func GetJobWithName(
	ctx context.Context,
	h *helper.Helper,
	name string,
	namespace string,
) (*batchv1.Job, error) {

	// Check if this Job already exists
	job := &batchv1.Job{}
	err := h.GetClient().Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, job)
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			return job, err
		}
		h.GetLogger().Info("GetJobWithName err")
		return job, err
	}

	return job, nil
}
