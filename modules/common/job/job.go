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

// NewJob returns an initialized Job.
func NewJob(
	job *batchv1.Job,
	jobType string,
	preserve bool,
	timeout time.Duration,
	beforeHash string,
) *Job {

	j := &Job{
		expectedJob: job,
		actualJob:   nil,
		jobType:     jobType,
		preserve:    preserve,
		timeout:     timeout,
		beforeHash:  beforeHash,
		changed:     false,
	}
	j.defaultTTL()
	return j
}

// createJob - creates job, reconciles after Xs if object won't exist.
func (j *Job) createJob(
	ctx context.Context,
	h *helper.Helper,
) (ctrl.Result, error) {
	job := &batchv1.Job{}
	job.ObjectMeta = j.expectedJob.ObjectMeta
	op, err := controllerutil.CreateOrPatch(ctx, h.GetClient(), job, func() error {
		job.Spec = j.expectedJob.Spec
		job.Annotations = util.MergeStringMaps(job.Annotations, map[string]string{hashAnnotationName: j.hash})
		err := controllerutil.SetControllerReference(h.GetBeforeObject(), job, h.GetScheme())
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			h.GetLogger().Info(fmt.Sprintf("Job %s not found, reconcile in %s", job.Name, j.timeout))
			return ctrl.Result{RequeueAfter: j.timeout}, nil
		}
		h.GetLogger().Error(err, "Job CreateOrPatch failed", "job", job.Name)
		return ctrl.Result{}, err
	}
	j.actualJob = job
	if op != controllerutil.OperationResultNone {
		h.GetLogger().Info(fmt.Sprintf("Job %s %s - %s", j.jobType, job.Name, op))
		return ctrl.Result{RequeueAfter: j.timeout}, nil
	}

	return ctrl.Result{}, nil
}

func (j *Job) defaultTTL() {
	// preserve has higher priority than having any kind of TTL
	if j.preserve {
		// so we reset TTL to avoid automatic deletion
		j.expectedJob.Spec.TTLSecondsAfterFinished = nil
		return
	}
	// if the client set a specific TTL then we honore it.
	if j.expectedJob.Spec.TTLSecondsAfterFinished != nil {
		return
	}
	// we are here as preserve is false and no TTL is set. We apply a default
	// TTL:
	// i) to make sure that the Job is eventually cleaned up
	// ii) to trigger the Job deletion with a delay to avoid racing between
	// Job deletion and callers reading old CR data from caches and re-creating
	// the Job. See more in https://github.com/openstack-k8s-operators/nova-operator/issues/110
	ttl := defaultTTL
	j.expectedJob.Spec.TTLSecondsAfterFinished = &ttl
}

// DoJob - run a job if the hashBefore and hash is different. If the job hash
// changes while the previous job still running then the first it waits for the
// previous job to finish then deletes the old job and runs the new one.
// (We do this as we assume that killing a job can leave the openstack
// deployment in an incosistent state.)
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

	// We intentionally only include the PodTemplate to the hash of the Job.
	// Fields outside of the PodTemplate, like TTL does not define what to run
	// just how to run it. So changing such field should not trigger the re-run
	// of the Job
	j.hash, err = util.ObjectHash(j.expectedJob.Spec.Template)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error calculating %s hash: %w", j.jobType, err)
	}

	if j.beforeHash != j.hash {
		j.changed = true
	}

	//
	// Check if this job already exists
	//
	j.actualJob, err = GetJobWithName(ctx, h, j.expectedJob.Name, j.expectedJob.Namespace)

	exists := !k8s_errors.IsNotFound(err)

	if err != nil && exists {
		return ctrl.Result{}, fmt.Errorf("error getting existing job %s : %w", j.jobType, err)
	}

	// If the hash of the job not changed then we don't need to create or wait
	// for any jobs
	if !j.changed {
		if exists {
			// but we  still want to allow changing the TTL on a finished job
			return j.updateTTL(ctx, h)
		}
		return ctrl.Result{}, nil
	}

	if exists {
		ctrlResult, err = j.waitOnJob(ctx, h)
		if err != nil || (ctrlResult != ctrl.Result{}) {
			return ctrlResult, err
		}
		// allow updating TTL even on running jobs
		ctrlResult, err = j.updateTTL(ctx, h)
		if err != nil || (ctrlResult != ctrl.Result{}) {
			return ctrlResult, err
		}
	} else {
		ctrlResult, err = j.createJob(ctx, h)
		if err != nil || (ctrlResult != ctrl.Result{}) {
			return ctrlResult, err
		}
	}

	return ctrl.Result{}, nil
}

func (j *Job) updateTTL(ctx context.Context, h *helper.Helper) (ctrl.Result, error) {
	job := &batchv1.Job{}
	job.ObjectMeta = j.expectedJob.ObjectMeta
	_, err := controllerutil.CreateOrPatch(ctx, h.GetClient(), job, func() error {
		job.Spec.TTLSecondsAfterFinished = j.expectedJob.Spec.TTLSecondsAfterFinished
		return nil
	})
	if err != nil {
		h.GetLogger().Info("Failed to update TTL on Job")
		return ctrl.Result{}, err
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

// DeleteJob deletes the batchv1.Job if exists. It is not an error to call
// this on an already deleted job.
func DeleteJob(
	ctx context.Context,
	h *helper.Helper,
	name string,
	namespace string,
) error {
	job := &batchv1.Job{}
	job.Name = name
	job.Namespace = namespace

	h.GetLogger().Info("Deleting Job", "Job.Namespace", namespace, "Job.Name", name)
	background := metav1.DeletePropagationBackground
	err := h.GetClient().Delete(ctx, job, &client.DeleteOptions{PropagationPolicy: &background})
	if err != nil && !k8s_errors.IsNotFound(err) {
		return err
	}

	return nil
}

func (j *Job) waitOnJob(
	ctx context.Context,
	h *helper.Helper,
) (ctrl.Result, error) {
	existingJobHash := j.actualJob.Annotations[hashAnnotationName]

	if j.actualJob.Status.Active > 0 {
		if existingJobHash != j.hash {
			h.GetLogger().Info(
				"The hash of the job changed while the job was running, " +
					"waiting for the previous job to finish before re-run.")
		}
		h.GetLogger().Info("Job Status Active... requeuing")
		return ctrl.Result{RequeueAfter: j.timeout}, nil
	} else if j.actualJob.Status.Succeeded > 0 {
		if existingJobHash != j.hash {
			h.GetLogger().Info(
				"The hash of the job changed but the previously succeeded job still exists. " +
					"Deleting old job and requeueing.")
			err := DeleteJob(ctx, h, j.actualJob.Name, j.actualJob.Namespace)
			if err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{RequeueAfter: j.timeout}, nil
		}
		h.GetLogger().Info("Job Status Successful")
		return ctrl.Result{}, nil
	} else if j.actualJob.Status.Failed > 0 {
		if existingJobHash != j.hash {
			h.GetLogger().Info(
				"The hash of the job changed but the previous failed job still exists. " +
					"Deleting old job and requeueing.")
			err := DeleteJob(ctx, h, j.actualJob.Name, j.actualJob.Namespace)
			if err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{RequeueAfter: j.timeout}, nil
		}
		h.GetLogger().Info("Job Status Failed")
		return ctrl.Result{}, k8s_errors.NewInternalError(errors.New("Job Failed. Check job logs"))
	} else {
		if existingJobHash != j.hash {
			h.GetLogger().Info(
				"The hash of the job changed while the job was incomplete, " +
					"waiting for the previous job to finish before re-run.")
		}
		h.GetLogger().Info("Job Status incomplete... requeuing")
		return ctrl.Result{RequeueAfter: j.timeout}, nil
	}
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
		h.GetLogger().Info("GetJobWithName %s err: %w", name, err)
		return job, err
	}

	return job, nil
}
