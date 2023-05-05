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
	timeout time.Duration, // unused
	beforeHash string,
) *Job {

	return &Job{
		job:        job,
		jobType:    jobType,
		preserve:   preserve,
		timeout:    timeout, // unused
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
			h.GetLogger().Info(fmt.Sprintf("Job %s not found", j.job.Name))
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, err
	}
	if op != controllerutil.OperationResultNone {
		h.GetLogger().Info(fmt.Sprintf("Job %s %s - %s", j.jobType, j.job.Name, op))
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func (j *Job) defaultTTL() {
	// preserve has higher priority than having any kind of TTL
	if j.preserve {
		// so we reset TTL to avoid automatic deletion
		j.job.Spec.TTLSecondsAfterFinished = nil
		return
	}
	// if the client set a specific TTL then we honore it.
	if j.job.Spec.TTLSecondsAfterFinished != nil {
		return
	}
	// we are here as preserve is false and no TTL is set. We apply a default
	// TTL:
	// i) to make sure that the Job is eventually cleaned up
	// ii) to trigger the Job deletion with a delay to avoid racing between
	// Job deletion and callers reading old CR data from caches and re-creating
	// the Job. See more in https://github.com/openstack-k8s-operators/nova-operator/issues/110
	ttl := defaultTTL
	j.job.Spec.TTLSecondsAfterFinished = &ttl
}

// DoJob - run a job if the hashBefore and hash is different. If there is an existing job, wait for the job
// to finish. Right now we do not expect the job to change while running.
// If TTLSecondsAfterFinished is unset on the Job and preserve is false, the Job
// will be deleted after 10 minutes. Set preserve to true if you want to keep
// the job, or set a specific value to job.Spec.TTLSecondsAfterFinished to
// define when the Job should be deleted.
func (j *Job) DoJob(
	ctx context.Context,
	h *helper.Helper,
) (ctrl.Result, error) {
	var ctrlResult ctrl.Result
	var err error

	j.hash, err = util.ObjectHash(j.job)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error calculating %s hash: %w", j.jobType, err)
	}

	// if the hash changed the job should run
	if j.beforeHash != j.hash {
		j.changed = true
	}

	// NOTE(gibi): This should be in NewJob but then the defaulting would affect
	// the hash of the Job calculated in DoJob above. As preserve can change
	// after the job finished and preserve is implemented by changing TTL the
	// change of preserve would change the hash of the Job after such hash is
	// persisted by the caller.
	// Moving hash calculation and defaultTTL to NewJob would be logically
	// possible but as hash calculation might fail NewJob inteface would need
	// to be changed to report the possible error.
	j.defaultTTL()

	//
	// Check if this job already exists
	//
	job, err := GetJobWithName(ctx, h, j.job.Name, j.job.Namespace)
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
		ctrlResult, err := waitOnJob(ctx, h, j.job.Name, j.job.Namespace)
		if (ctrlResult != ctrl.Result{}) {
			return ctrlResult, nil
		}
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	// allow updating TTLSecondsAfterFinished even after the job is finished
	job, err = GetJobWithName(ctx, h, j.job.Name, j.job.Namespace)
	if err != nil && !k8s_errors.IsNotFound(err) {
		return ctrl.Result{}, err
	}

	if err == nil {
		_, err = controllerutil.CreateOrPatch(ctx, h.GetClient(), job, func() error {
			job.Spec.TTLSecondsAfterFinished = j.job.Spec.TTLSecondsAfterFinished
			return nil
		})

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

// waitOnJob func -  returns true if the job
func waitOnJob(
	ctx context.Context,
	h *helper.Helper,
	name string,
	namespace string,
) (ctrl.Result, error) {
	// Check if this Job already exists
	job, err := GetJobWithName(ctx, h, name, namespace)
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			h.GetLogger().Info("Job was not found.")
			return ctrl.Result{}, err
		}
		h.GetLogger().Info("WaitOnJob err")
		return ctrl.Result{}, err
	}

	if job.Status.Active > 0 {
		h.GetLogger().Info("Job Status Active... returning")
		return ctrl.Result{}, nil
	} else if job.Status.Succeeded > 0 {
		h.GetLogger().Info("Job Status Successful")
		return ctrl.Result{}, nil
	} else if job.Status.Failed > 0 {
		h.GetLogger().Info("Job Status Failed")
		return ctrl.Result{}, k8s_errors.NewInternalError(errors.New("Job Failed. Check job logs"))
	}
	h.GetLogger().Info("Job Status incomplete... returning")
	return ctrl.Result{}, nil
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
