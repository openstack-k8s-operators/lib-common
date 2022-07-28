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
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"errors"
)

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
		timeout:    timeout, // timeout to set in s to reconcile
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

		return nil
	})
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			h.GetLogger().Info(fmt.Sprintf("Job %s not found, reconcile in %ds", j.job.Name, j.timeout))
			return ctrl.Result{RequeueAfter: time.Duration(j.timeout) * time.Second}, nil
		}
		return ctrl.Result{}, err
	}
	if op != controllerutil.OperationResultNone {
		h.GetLogger().Info(fmt.Sprintf("Job %s %s - %s", j.jobType, j.job.Name, op))
		return ctrl.Result{RequeueAfter: time.Duration(j.timeout) * time.Second}, nil
	}

	return ctrl.Result{}, nil
}

//
// DoJob - run a job if the hashBefore and hash is different. If there is an existing job, wait for the job
// to finish. Right now we do not expect the job to change while running. If the job finished successful
// and preserve flag is not set it gets deleted.
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
		ctrlResult, err := WaitOnJob(ctx, h, j.job.Name, j.job.Namespace, j.timeout)
		if (ctrlResult != ctrl.Result{}) {
			return ctrlResult, nil
		}
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	// delete the job if PreserveJobs is not enabled
	if !j.preserve {
		err = DeleteJob(ctx, h, j.job.Name, j.job.Namespace)
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
		h.GetLogger().Info("Deleting Job", "Job.Namespace", namespace, "Job.Name", name)
		background := metav1.DeletePropagationBackground
		err = h.GetKClient().BatchV1().Jobs(foundJob.Namespace).Delete(
			ctx, foundJob.Name, metav1.DeleteOptions{PropagationPolicy: &background})
		if err != nil {
			return err
		}
		return err
	}
	return nil
}

// WaitOnJob func -  returns true if the job
func WaitOnJob(
	ctx context.Context,
	h *helper.Helper,
	name string,
	namespace string,
	timeout int,
) (ctrl.Result, error) {
	// Check if this Job already exists
	job, err := GetJobWithName(ctx, h, name, namespace)
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			h.GetLogger().Info("Job was not found.")
			return ctrl.Result{RequeueAfter: time.Second * time.Duration(timeout)}, nil
		}
		h.GetLogger().Info("WaitOnJob err")
		return ctrl.Result{}, err
	}

	if job.Status.Active > 0 {
		h.GetLogger().Info("Job Status Active... requeuing")
		return ctrl.Result{RequeueAfter: time.Second * time.Duration(timeout)}, nil
	} else if job.Status.Failed > 0 {
		h.GetLogger().Info("Job Status Failed")
		return ctrl.Result{}, k8s_errors.NewInternalError(errors.New("Job Failed. Check job logs"))
	} else if job.Status.Succeeded > 0 {
		h.GetLogger().Info("Job Status Successful")
		return ctrl.Result{}, nil
	}
	h.GetLogger().Info("Job Status incomplete... requeuing")
	return ctrl.Result{RequeueAfter: time.Second * time.Duration(timeout)}, nil
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
