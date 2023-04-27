/*
Copyright 2023 Red Hat

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
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
)

const (
	timeout  = time.Duration(1) * time.Second
	noHash   = ""
	preserve = true
)

type MockJobHelper struct {
	Job              *batchv1.Job
	ControllerRefSet bool
}

func (m *MockJobHelper) GetBeforeObject() client.Object {
	return &unstructured.Unstructured{}
}

func (m *MockJobHelper) GetScheme() *runtime.Scheme {
	return nil
}

func (m *MockJobHelper) GetLogger() logr.Logger {
	return ctrl.Log.WithName("test")
}

func (m *MockJobHelper) CreateOrPatch(ctx context.Context, obj client.Object, f controllerutil.MutateFn) (controllerutil.OperationResult, error) {
	op := controllerutil.OperationResultNone
	if m.Job == nil {
		op = controllerutil.OperationResultCreated
	} else {
		op = controllerutil.OperationResultUpdated
		obj.(*batchv1.Job).Spec = m.Job.Spec
		obj.(*batchv1.Job).Status = m.Job.Status
	}

	f()

	m.Job = obj.(*batchv1.Job)

	return op, nil
}

func (m *MockJobHelper) SetControllerReference(owner, controlled metav1.Object, scheme *runtime.Scheme) error {
	m.ControllerRefSet = true
	return nil
}

func (m *MockJobHelper) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	if m.Job == nil {
		return k8s_errors.NewNotFound(schema.GroupResource{}, obj.GetName())
	}
	obj.(*batchv1.Job).Spec = m.Job.Spec
	obj.(*batchv1.Job).Status = m.Job.Status
	return nil

}

func getExampleJob() *batchv1.Job {
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-job",
			Namespace: "test-namespace",
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyOnFailure,
					Containers: []corev1.Container{
						{
							Name: "test-job-pod",
							Command: []string{
								"/bin/bash",
							},
							Image: "test-container-image",
						},
					},
				},
			},
		},
	}

}

func TestDoJobTTLDefaultedWhenNotProvided(t *testing.T) {
	g := NewWithT(t)
	ctx := context.TODO()
	job := getExampleJob()
	j := NewJob(job, "test-job", !preserve, time.Duration(1)*time.Second, noHash)
	mock := &MockJobHelper{}

	_, err := j.DoJob(ctx, mock)

	g.Expect(err).ShouldNot(HaveOccurred())
	gotJob, err := GetJobWithName(ctx, mock, job.Name, job.Namespace)
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(*gotJob.Spec.TTLSecondsAfterFinished).To(Equal(defaultTTL))
}

func TestDoJobTTLKeptIfProvided(t *testing.T) {
	g := NewWithT(t)
	ctx := context.TODO()
	job := getExampleJob()
	var ttl int32 = 13
	job.Spec.TTLSecondsAfterFinished = &ttl
	j := NewJob(job, "test-job", !preserve, time.Duration(1)*time.Second, noHash)
	mock := &MockJobHelper{}

	_, err := j.DoJob(ctx, mock)

	g.Expect(err).ShouldNot(HaveOccurred())
	gotJob, err := GetJobWithName(ctx, mock, job.Name, job.Namespace)
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(*gotJob.Spec.TTLSecondsAfterFinished).To(Equal(ttl))
}

func TestDoJobTTLSetToInfiniteIfPreserveRequested(t *testing.T) {
	g := NewWithT(t)
	ctx := context.TODO()
	job := getExampleJob()
	var ttl int32 = 13
	job.Spec.TTLSecondsAfterFinished = &ttl
	j := NewJob(job, "test-job", preserve, time.Duration(1)*time.Second, noHash)
	mock := &MockJobHelper{}

	_, err := j.DoJob(ctx, mock)

	g.Expect(err).ShouldNot(HaveOccurred())
	gotJob, err := GetJobWithName(ctx, mock, job.Name, job.Namespace)
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(gotJob.Spec.TTLSecondsAfterFinished).To(BeNil())
}

func TestDoJobNoJobExistsRunToCompletion(t *testing.T) {
	g := NewWithT(t)
	ctx := context.TODO()
	job := getExampleJob()
	j := NewJob(job, "test-job", false, time.Duration(1)*time.Second, noHash)
	mock := &MockJobHelper{}
	result, err := j.DoJob(ctx, mock)

	// The caller is asked to requeue as the job is not finished yet
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(result).To(Equal(ctrl.Result{RequeueAfter: timeout}))

	// Job is created in the backend with a controller ref
	g.Expect(mock.Job).NotTo(BeNil())
	g.Expect(mock.ControllerRefSet).To(BeTrue())

	// The passed in hash, that was empty, is different from the hash of the
	// job
	g.Expect(j.HasChanged()).To(BeTrue())
	g.Expect(j.GetHash()).NotTo(Equal(noHash))

	// Simulate that the Job succeeded
	mock.Job.Status.Succeeded = 1

	result, err = j.DoJob(ctx, mock)

	// The empty result signals the caller that the job is finished
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(result).To(Equal(ctrl.Result{}))
	// the hash is still different from the empty hash
	g.Expect(j.HasChanged()).To(BeTrue())
	g.Expect(j.GetHash()).NotTo(Equal(noHash))

}

func TestDoJobNewHashOldJobStillExists(t *testing.T) {
	g := NewWithT(t)
	ctx := context.TODO()
	job := getExampleJob()
	j := NewJob(job, "test-job", false, time.Duration(1)*time.Second, noHash)
	mock := &MockJobHelper{}

	// run a job to completion
	_, err := j.DoJob(ctx, mock)
	g.Expect(err).ShouldNot(HaveOccurred())
	mock.Job.Status.Succeeded = 1
	result, err := j.DoJob(ctx, mock)
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(result).To(Equal(ctrl.Result{}))

	// store the job's hash after it is finished
	storedHash := j.GetHash()
	g.Expect(storedHash).NotTo(BeEmpty())

	// requests a new job as the input of the job is changed, e.g. the image of
	// the job is changed
	job = getExampleJob()
	job.Spec.Template.Spec.Containers[0].Image = "new-image"
	j = NewJob(job, "test-job", false, time.Duration(1)*time.Second, storedHash)
	result, err = j.DoJob(ctx, mock)
	g.Expect(err).ShouldNot(HaveOccurred())
	// BUG
	// We expect a new job being created as the existing one has a different
	// hash than the newly requested one and therefore DoJob should request a
	// requeue to wait for the Job to finish.
	//
	// g.Expect(result).To(Equal(ctrl.Result{RequeueAfter: timeout}))
	//
	// But instead the lib-common code sees the previous finished job and
	// returns that the job is succeeded, but now with the new hash without
	// creating a new job.
	g.Expect(result).To(Equal(ctrl.Result{}))

	g.Expect(j.HasChanged()).To(BeTrue())
	g.Expect(j.GetHash()).NotTo(Equal(storedHash))
}

func TestDoJobNewHashOldJobAlreadyDeleted(t *testing.T) {
	g := NewWithT(t)
	ctx := context.TODO()
	job := getExampleJob()
	j := NewJob(job, "test-job", false, time.Duration(1)*time.Second, noHash)
	mock := &MockJobHelper{}

	// run a job to completion
	_, err := j.DoJob(ctx, mock)
	g.Expect(err).ShouldNot(HaveOccurred())
	mock.Job.Status.Succeeded = 1
	result, err := j.DoJob(ctx, mock)
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(result).To(Equal(ctrl.Result{}))

	// store the job's hash after it is finished
	storedHash := j.GetHash()
	g.Expect(storedHash).NotTo(BeEmpty())

	// simulate that the TTL of the job is expired and therefore the job is
	// deleted

	mock.Job = nil

	// requests a new job as the input of the job is changed, e.g. the image of
	// the job is changed
	job = getExampleJob()
	job.Spec.Template.Spec.Containers[0].Image = "new-image"
	j = NewJob(job, "test-job", false, time.Duration(1)*time.Second, storedHash)
	result, err = j.DoJob(ctx, mock)

	// we expect that a new job is created and the client is requested to
	// requeue while waiting for the job to finish
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(result).To(Equal(ctrl.Result{RequeueAfter: timeout}))
	g.Expect(j.HasChanged()).To(BeTrue())
	g.Expect(j.GetHash()).NotTo(Equal(storedHash))
}
