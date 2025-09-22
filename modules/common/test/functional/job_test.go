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
	. "github.com/onsi/ginkgo/v2" // nolint:revive
	. "github.com/onsi/gomega"    // nolint:revive
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

var (
	requeue  = ctrl.Result{RequeueAfter: timeout}
	finished = ctrl.Result{}
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

func runJobSuccessfully(namespace string) (*job.Job, *batchv1.Job) {
	exampleJob := getExampleJob(namespace)
	j := job.NewJob(exampleJob, "test-job", !preserve, timeout, noHash)

	result, err := j.DoJob(ctx, h)

	// The caller is asked to requeue as the job is not finished yet
	Expect(err).ShouldNot(HaveOccurred())
	Expect(result).To(Equal(requeue))

	// a k8s Job is created with an controller reference
	k8sJob := th.GetJob(th.GetName(exampleJob))
	Expect(k8sJob.GetOwnerReferences()).To(HaveLen(1))
	Expect(k8sJob.GetOwnerReferences()[0]).To(HaveField("Name", h.GetBeforeObject().GetName()))
	t := true
	Expect(k8sJob.GetOwnerReferences()[0]).To(HaveField("Controller", &t))

	// Simulate that the Job succeeded
	th.SimulateJobSuccess(th.GetName(exampleJob))

	result, err = j.DoJob(ctx, h)

	// The empty result signals the caller that the job is finished
	Expect(err).ShouldNot(HaveOccurred())
	Expect(result).To(Equal(finished))
	Expect(j.HasChanged()).To(BeTrue())
	Expect(j.GetHash()).NotTo(Equal(noHash))

	return j, th.GetJob(th.GetName(exampleJob))
}

var _ = Describe("job package", func() {
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

	It("TTL can be updated after the job is finished", func() {
		exampleJob := getExampleJob(namespace)
		j := job.NewJob(exampleJob, "test-job", !preserve, timeout, noHash)

		_, err := j.DoJob(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())
		th.SimulateJobSuccess(th.GetName(exampleJob))
		result, err := j.DoJob(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(result).To(Equal(finished))

		k8sJob := th.GetJob(th.GetName(exampleJob))
		Expect(*k8sJob.Spec.TTLSecondsAfterFinished).To(Equal(int32(600)))

		var newTTL int32 = 13
		exampleJob.Spec.TTLSecondsAfterFinished = &newTTL
		result, err = j.DoJob(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(result).To(Equal(finished))

		k8sJob = th.GetJob(th.GetName(exampleJob))
		Expect(*k8sJob.Spec.TTLSecondsAfterFinished).To(Equal(newTTL))
	})

	It("does not re-create a job if only TTL changes", func() {
		j, k8sJob := runJobSuccessfully(namespace)
		successfulJobHash := j.GetHash()

		background := metav1.DeletePropagationBackground
		th.DeleteInstance(k8sJob, &client.DeleteOptions{PropagationPolicy: &background})

		// We expect that the TTL change is accepted but as the job is
		// already deleted no new job is created just for the TTL change
		newJob := getExampleJob(namespace)
		var newTTL int32 = 13
		newJob.Spec.TTLSecondsAfterFinished = &newTTL
		j = job.NewJob(newJob, "test-job", !preserve, timeout, successfulJobHash)
		result, err := j.DoJob(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(result).To(Equal(finished))
		err = cClient.Get(ctx, th.GetName(k8sJob), k8sJob)
		Expect(k8s_errors.IsNotFound(err)).To(BeTrue())
	})

	It("runs the job if it has a new hash and the job does not exists", func() {
		runJobSuccessfully(namespace)
	})

	It("re-runs the job if its hash differs and the previous job exists", func() {
		j, k8sJob := runJobSuccessfully(namespace)
		// store the job's hash after it is finished
		storedHash := j.GetHash()
		Expect(storedHash).NotTo(BeEmpty())

		// requests a new job as the input of the job is changed, e.g. the image of
		// the job is changed
		newJob := getExampleJob(namespace)
		newJob.Spec.Template.Spec.Containers[0].Image = "new-image"
		j = job.NewJob(newJob, "test-job", !preserve, timeout, noHash)
		result, err := j.DoJob(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())
		// We expect that the old job is deleted and DoJob request a requeue
		// so that the next DoJob call can create a new Job
		Expect(result).To(Equal(requeue))

		result, err = j.DoJob(ctx, h)
		// now there is a new job created, but requeue is still requested as
		// it is not succeeded yet.
		Expect(err).ShouldNot(HaveOccurred())
		Expect(result).To(Equal(requeue))
		Expect(th.GetJob(th.GetName(newJob)).UID).NotTo(Equal(k8sJob.UID))

		th.SimulateJobSuccess(th.GetName(newJob))

		// Now the job is finished with a new hash
		result, err = j.DoJob(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(result).To(Equal(finished))
		Expect(j.HasChanged()).To(BeTrue())
		Expect(j.GetHash()).NotTo(Equal(storedHash))
	})

	It("re-runs the job if its hash differs and the previous already deleted", func() {
		j, k8sJob := runJobSuccessfully(namespace)
		// store the job's hash after it is finished
		storedHash := j.GetHash()
		Expect(storedHash).NotTo(BeEmpty())

		// simulate that the TTL of the job is expired and therefore the job is
		// deleted
		// need background propagation policy otherwise the Job remains
		// in orphan state
		background := metav1.DeletePropagationBackground
		th.DeleteInstance(k8sJob, &client.DeleteOptions{PropagationPolicy: &background})

		// requests a new job as the input of the job is changed, e.g. the image of
		// the job is changed
		newJob := getExampleJob(namespace)
		newJob.Spec.Template.Spec.Containers[0].Image = "new-image"
		j = job.NewJob(newJob, "test-job", !preserve, timeout, noHash)
		result, err := j.DoJob(ctx, h)
		// we expect that a new job is created and the client is requested to
		// requeue while waiting for the job to finish
		Expect(err).ShouldNot(HaveOccurred())
		Expect(result).To(Equal(requeue))
		Expect(j.HasChanged()).To(BeTrue())
		Expect(j.GetHash()).NotTo(Equal(storedHash))

		Expect(th.GetJob(th.GetName(newJob)).UID).NotTo(Equal(k8sJob.UID))
	})

	It("reports failure if the job failed", func() {
		exampleJob := getExampleJob(namespace)
		j := job.NewJob(exampleJob, "test-job", !preserve, timeout, noHash)

		result, err := j.DoJob(ctx, h)

		// The caller is asked to requeue as the job is not finished yet
		Expect(err).ShouldNot(HaveOccurred())
		Expect(result).To(Equal(requeue))

		// a k8s Job is created in with an controller reference
		th.GetJob(th.GetName(exampleJob))

		// Simulate that the Job succeeded
		th.SimulateJobFailure(th.GetName(exampleJob))

		_, err = j.DoJob(ctx, h)

		Expect(err).Should(HaveOccurred())
		var statusErr *k8s_errors.StatusError
		Expect(errors.As(err, &statusErr)).To(BeTrue())
		Expect(statusErr.Status().Message).To(ContainSubstring("Check job logs"))
	})

	It("requeue if the job definition is changed while the old job still running and the wait for the old job to finish before re-run", func() {
		exampleJob := getExampleJob(namespace)
		j := job.NewJob(exampleJob, "test-job", !preserve, timeout, noHash)

		result, err := j.DoJob(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(result).To(Equal(requeue))
		oldHash := j.GetHash()

		newJob := getExampleJob(namespace)
		newJob.Spec.Template.Spec.Containers[0].Image = "new-image"

		j = job.NewJob(newJob, "test-job", !preserve, timeout, noHash)
		result, err = j.DoJob(ctx, h)
		// As the current job still not finished it requests requeu until it
		// finishes
		Expect(err).ShouldNot(HaveOccurred())
		Expect(result).To(Equal(requeue))

		// Assert that the running job is not changed so it still has the
		// image of the original request
		oldK8sJob := th.GetJob(th.GetName(newJob))
		Expect(oldK8sJob.Spec.Template.Spec.Containers[0].Image).To(
			Equal(exampleJob.Spec.Template.Spec.Containers[0].Image))

		// Simulate that the original Job succeeds
		th.SimulateJobSuccess(th.GetName(exampleJob))

		// We expect that if DoJob is called with the new job now then the old
		// job is deleted and requeue is requested
		result, err = j.DoJob(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(result).To(Equal(requeue))

		// at the next DoJob a new job is created with the new content but
		// requeue is still requested as the new job not finished yet
		result, err = j.DoJob(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(result).To(Equal(requeue))
		newK8sJob := th.GetJob(th.GetName(newJob))
		Expect(newK8sJob.UID).NotTo(Equal(oldK8sJob.UID))
		Expect(newK8sJob.Spec.Template.Spec.Containers[0].Image).To(
			Equal(newJob.Spec.Template.Spec.Containers[0].Image))

		th.SimulateJobSuccess(th.GetName(newJob))
		// now the job finished with the new hash
		result, err = j.DoJob(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(result).To(Equal(finished))
		Expect(j.GetHash()).NotTo(Equal(oldHash))
	})

	It("deletes the failed job if hash is changed and re-runs", func() {
		exampleJob := getExampleJob(namespace)
		j := job.NewJob(exampleJob, "test-job", !preserve, timeout, noHash)

		result, err := j.DoJob(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(result).To(Equal(requeue))
		oldHash := j.GetHash()

		newJob := getExampleJob(namespace)
		newJob.Spec.Template.Spec.Containers[0].Image = "new-image"

		j = job.NewJob(newJob, "test-job", !preserve, timeout, noHash)
		result, err = j.DoJob(ctx, h)
		// As the current job still not finished it requests requeu until it
		// finishes
		Expect(err).ShouldNot(HaveOccurred())
		Expect(result).To(Equal(requeue))
		oldK8sJob := th.GetJob(th.GetName(newJob))

		// Simulate that the original Job fails
		th.SimulateJobFailure(th.GetName(exampleJob))

		// We expect that if DoJob is called with the new job now then the old
		// job is deleted and requeue is requested
		result, err = j.DoJob(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(result).To(Equal(requeue))

		// at the next DoJob a new job is created with the new content but
		// requeue is still requested as the new job not finished yet
		result, err = j.DoJob(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(result).To(Equal(requeue))
		newK8sJob := th.GetJob(th.GetName(newJob))
		Expect(newK8sJob.UID).NotTo(Equal(oldK8sJob.UID))
		Expect(newK8sJob.Spec.Template.Spec.Containers[0].Image).To(
			Equal(newJob.Spec.Template.Spec.Containers[0].Image))

		th.SimulateJobSuccess(th.GetName(newJob))
		// now the job finished with the new hash
		result, err = j.DoJob(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(result).To(Equal(finished))
		Expect(j.GetHash()).NotTo(Equal(oldHash))
	})

	It("DeleteJob deletes existing jobs", func() {
		_, k8sJob := runJobSuccessfully(namespace)
		// the job exists
		th.GetJob(th.GetName(k8sJob))

		// assert that DeleteJob deletes it properly so the k8sJob not found
		// any more
		Expect(job.DeleteJob(ctx, h, k8sJob.Name, namespace)).To(Succeed())
		Eventually(func(g Gomega) {
			err := cClient.Get(ctx, th.GetName(k8sJob), k8sJob)
			g.Expect(err).To(HaveOccurred())
			var statusErr *k8s_errors.StatusError
			g.Expect(errors.As(err, &statusErr)).To(BeTrue())
			g.Expect(statusErr.Status().Reason).To(Equal(metav1.StatusReasonNotFound))
		}, timeout, interval).Should(Succeed())
	})

	It("DeleteJob ignores if the job is already deleted", func() {
		Expect(job.DeleteJob(ctx, h, "non-existent-job", namespace)).To(Succeed())
	})

})
