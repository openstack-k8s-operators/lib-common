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
package functional

import (
	"errors"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openstack-k8s-operators/lib-common/modules/common/job"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	noHash   = ""
	preserve = true
)

func getExampleJob(namespace string) *batchv1.Job {
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-job",
			Namespace: namespace,
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

var _ = Describe("job.Job", func() {
	var namespace string

	BeforeEach(func() {
		// NOTE(gibi): We need to create a unique namespace for each test run
		// as namespaces cannot be deleted in a locally running envtest. See
		// https://book.kubebuilder.io/reference/envtest.html#namespace-usage-limitation
		namespace = uuid.New().String()
		th.CreateNamespace(namespace)
		// We still request the delete of the Namespace to properly cleanup if
		// we run the test in an existing cluster.
		DeferCleanup(th.DeleteNamespace, namespace)

	})

	It("defaults the TTL if not provided", func() {
		exampleJob := getExampleJob(namespace)
		j := job.NewJob(exampleJob, "test-job", !preserve, timeout, noHash)

		_, err := j.DoJob(ctx, h)

		Expect(err).ShouldNot(HaveOccurred())
		gotJob, err := job.GetJobWithName(ctx, h, exampleJob.Name, namespace)
		Expect(err).ShouldNot(HaveOccurred())
		// job.defaultTTL is 600
		Expect(*gotJob.Spec.TTLSecondsAfterFinished).To(Equal(int32(600)))
	})

	It("keeps the requested TTL", func() {
		exampleJob := getExampleJob(namespace)
		var ttl int32 = 13
		exampleJob.Spec.TTLSecondsAfterFinished = &ttl
		j := job.NewJob(exampleJob, "test-job", !preserve, timeout, noHash)

		_, err := j.DoJob(ctx, h)

		Expect(err).ShouldNot(HaveOccurred())
		gotJob, err := job.GetJobWithName(ctx, h, exampleJob.Name, namespace)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(*gotJob.Spec.TTLSecondsAfterFinished).To(Equal(ttl))
	})

	It("sets the TTL to infinite if preserve is requested", func() {
		exampleJob := getExampleJob(namespace)
		var ttl int32 = 13
		exampleJob.Spec.TTLSecondsAfterFinished = &ttl
		j := job.NewJob(exampleJob, "test-job", preserve, timeout, noHash)

		_, err := j.DoJob(ctx, h)

		Expect(err).ShouldNot(HaveOccurred())
		gotJob, err := job.GetJobWithName(ctx, h, exampleJob.Name, namespace)
		Expect(err).ShouldNot(HaveOccurred())
		// TTL=nil means TTL is infinite
		Expect(gotJob.Spec.TTLSecondsAfterFinished).To(BeNil())
	})

	It("runs the job if it has a new hash and the job does not exists", func() {
		exampleJob := getExampleJob(namespace)
		j := job.NewJob(exampleJob, "test-job", !preserve, timeout, noHash)

		result, err := j.DoJob(ctx, h)

		// The caller is asked to requeue as the job is not finished yet
		Expect(err).ShouldNot(HaveOccurred())
		Expect(result).To(Equal(ctrl.Result{RequeueAfter: timeout}))

		// a k8s Job is created in with an controller reference
		k8sJob := th.GetJob(th.GetName(exampleJob))
		Expect(k8sJob.GetOwnerReferences()).To(HaveLen(1))
		Expect(k8sJob.GetOwnerReferences()[0]).To(HaveField("Name", h.GetBeforeObject().GetName()))
		t := true
		Expect(k8sJob.GetOwnerReferences()[0]).To(HaveField("Controller", &t))

		// The passed in hash, that was empty, is different from the hash of the
		// job
		Expect(j.HasChanged()).To(BeTrue())
		Expect(j.GetHash()).NotTo(Equal(noHash))

		// Simulate that the Job succeeded
		th.SimulateJobSuccess(th.GetName(exampleJob))

		result, err = j.DoJob(ctx, h)

		// The empty result signals the caller that the job is finished
		Expect(err).ShouldNot(HaveOccurred())
		Expect(result).To(Equal(ctrl.Result{}))
		// the hash is still different from the empty hash
		Expect(j.HasChanged()).To(BeTrue())
		Expect(j.GetHash()).NotTo(Equal(noHash))
	})

	It("re-runs the job if its hash differs and the previous job exists", func() {
		exampleJob := getExampleJob(namespace)
		j := job.NewJob(exampleJob, "test-job", !preserve, timeout, noHash)

		// runs the job to completion
		_, err := j.DoJob(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())
		th.SimulateJobSuccess(th.GetName(exampleJob))
		result, err := j.DoJob(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(result).To(Equal(ctrl.Result{}))

		// store the job's hash after it is finished
		storedHash := j.GetHash()
		Expect(storedHash).NotTo(BeEmpty())

		k8sJobUID := th.GetJob(th.GetName(exampleJob)).UID

		// requests a new job as the input of the job is changed, e.g. the image of
		// the job is changed
		newJob := getExampleJob(namespace)
		newJob.Spec.Template.Spec.Containers[0].Image = "new-image"
		j = job.NewJob(newJob, "test-job", !preserve, timeout, noHash)
		result, err = j.DoJob(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())
		// BUG
		// We expect a new job being created as the existing one has a different
		// hash than the newly requested one and therefore DoJob should request a
		// requeue to wait for the Job to finish.
		//
		// g.Expect(result).To(Equal(ctrl.Result{RequeueAfter: timeout}))
		//
		// But instead the lib-common code sees the previous finished k8s job
		// and returns that the job is succeeded, but now with the new hash
		// without creating a new job.
		Expect(result).To(Equal(ctrl.Result{}))

		Expect(j.HasChanged()).To(BeTrue())
		Expect(j.GetHash()).NotTo(Equal(storedHash))

		// This proves that there was no new k8s Job created with the same
		// name as the UID is the same as the first job.
		Expect(th.GetJob(th.GetName(newJob)).UID).To(Equal(k8sJobUID))
	})

	It("re-runs the job if its hash differs and the previous already deleted", func() {
		exampleJob := getExampleJob(namespace)
		j := job.NewJob(exampleJob, "test-job", !preserve, timeout, noHash)

		// runs the job to completion
		_, err := j.DoJob(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())
		th.SimulateJobSuccess(th.GetName(exampleJob))
		result, err := j.DoJob(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(result).To(Equal(ctrl.Result{}))

		// store the job's hash after it is finished
		storedHash := j.GetHash()
		Expect(storedHash).NotTo(BeEmpty())

		k8sJobUID := th.GetJob(th.GetName(exampleJob)).UID

		// simulate that the TTL of the job is expired and therefore the job is
		// deleted
		// need background propagation policy otherwise the Job remains
		// in orphan state
		background := metav1.DeletePropagationBackground
		th.DeleteInstance(exampleJob, &client.DeleteOptions{PropagationPolicy: &background})

		// requests a new job as the input of the job is changed, e.g. the image of
		// the job is changed
		newJob := getExampleJob(namespace)
		newJob.Spec.Template.Spec.Containers[0].Image = "new-image"
		j = job.NewJob(newJob, "test-job", !preserve, timeout, noHash)
		result, err = j.DoJob(ctx, h)
		// we expect that a new job is created and the client is requested to
		// requeue while waiting for the job to finish
		Expect(err).ShouldNot(HaveOccurred())
		Expect(result).To(Equal(ctrl.Result{RequeueAfter: timeout}))

		Expect(j.HasChanged()).To(BeTrue())
		Expect(j.GetHash()).NotTo(Equal(storedHash))

		Expect(th.GetJob(th.GetName(newJob)).UID).NotTo(Equal(k8sJobUID))
	})

	It("reports failure if the job failed", func() {
		exampleJob := getExampleJob(namespace)
		j := job.NewJob(exampleJob, "test-job", !preserve, timeout, noHash)

		result, err := j.DoJob(ctx, h)

		// The caller is asked to requeue as the job is not finished yet
		Expect(err).ShouldNot(HaveOccurred())
		Expect(result).To(Equal(ctrl.Result{RequeueAfter: timeout}))

		// a k8s Job is created in with an controller reference
		th.GetJob(th.GetName(exampleJob))

		// Simulate that the Job succeeded
		th.SimulateJobFailure(th.GetName(exampleJob))

		_, err = j.DoJob(ctx, h)

		Expect(err).Should(HaveOccurred())
		var statusErr *k8s_errors.StatusError
		Expect(errors.As(err, &statusErr)).To(BeTrue())
		Expect(statusErr.Status().Message).To(ContainSubstring("Job Failed"))
	})

	It("reports error if the job definition is changed while the job still running", func() {
		exampleJob := getExampleJob(namespace)
		j := job.NewJob(exampleJob, "test-job", !preserve, timeout, noHash)

		result, err := j.DoJob(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(result).NotTo(Equal(ctrl.Result{}))
		oldHash := j.GetHash()

		newJob := getExampleJob(namespace)
		newJob.Spec.Template.Spec.Containers[0].Image = "new-image"

		j = job.NewJob(newJob, "test-job", !preserve, timeout, noHash)
		result, err = j.DoJob(ctx, h)
		// BUG: we should report an error to the caller to indicate that
		// the job cannot be changed as it is still running
		Expect(err).ShouldNot(HaveOccurred())
		Expect(result).NotTo(Equal(ctrl.Result{}))

		// Assert that the running job is not changed so it still has the
		// image of the original request
		k8sJob := th.GetJob(th.GetName(newJob))
		Expect(k8sJob.Spec.Template.Spec.Containers[0].Image).To(
			Equal(exampleJob.Spec.Template.Spec.Containers[0].Image))

		// Simulate that the original Job succeeds
		th.SimulateJobSuccess(th.GetName(exampleJob))

		// We expect that if DoJob is called with the new job now then a
		// new k8s job is created to re-run the job
		// BUG: but instead we report success for the old job content
		// with the new jobs hash
		result, err = j.DoJob(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(result).To(Equal(ctrl.Result{}))
		// the job content is not changed
		k8sJob = th.GetJob(th.GetName(newJob))
		Expect(k8sJob.Spec.Template.Spec.Containers[0].Image).To(
			Equal(exampleJob.Spec.Template.Spec.Containers[0].Image))
		// but the reported job has does
		Expect(j.GetHash()).NotTo(Equal(oldHash))
	})

})
