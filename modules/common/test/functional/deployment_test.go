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
	"fmt"
	"time"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openstack-k8s-operators/lib-common/modules/common/deployment"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

const (
	replicas int32 = 3
)

func getExampleDeployment(namespace string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment",
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: ptr.To(replicas),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"foo": "bar",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"foo": "bar",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "test-depl-pod",
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

func runDeploymentSuccessfully(namespace string) (*deployment.Deployment, *appsv1.Deployment) {
	exampleDepl := getExampleDeployment(namespace)
	d := deployment.NewDeployment(exampleDepl, timeout)

	_, err := d.CreateOrPatch(ctx, h)
	Expect(err).ShouldNot(HaveOccurred())

	// a k8s Deployment is created with an controller reference
	k8sDepl := d.GetDeployment()
	Expect(k8sDepl.GetOwnerReferences()).To(HaveLen(1))
	Expect(k8sDepl.GetOwnerReferences()[0]).To(HaveField("Name", h.GetBeforeObject().GetName()))
	t := true
	Expect(k8sDepl.GetOwnerReferences()[0]).To(HaveField("Controller", &t))

	// Simulate that the Deployment replicas are ready
	th.SimulateDeploymentReplicaReady(th.GetName(exampleDepl))

	return d, th.GetDeployment(th.GetName(exampleDepl))
}

var _ = Describe("deployment package", func() {
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

	It("defaults the poll interval and timeout if not explicite set", func() {
		exampleDepl := getExampleDeployment(namespace)
		d := deployment.NewDeployment(exampleDepl, timeout)

		// defaults PollInterval
		Expect(*d.GetRolloutPollInterval()).To(Equal(deployment.DefaultPollInterval))
		Expect(*d.GetRolloutPollTimeout()).To(Equal(deployment.DefaultPollTimeout))

		_, err := d.CreateOrPatch(ctx, h)

		Expect(err).ShouldNot(HaveOccurred())
		_, err = deployment.GetDeploymentWithName(ctx, h, exampleDepl.Name, namespace)
		Expect(err).ShouldNot(HaveOccurred())
	})

	It("use custom poll timeout/interval when set", func() {
		exampleDepl := getExampleDeployment(namespace)
		d := deployment.NewDeployment(exampleDepl, timeout)

		// custom poll settings
		customInterval := 10 * time.Second
		customTimeout := 60 * time.Second

		d.SetRolloutPollInterval(customInterval)
		d.SetRolloutPollTimeout(customTimeout)

		Expect(*d.GetRolloutPollInterval()).To(Equal(customInterval))
		Expect(*d.GetRolloutPollTimeout()).To(Equal(customTimeout))
	})

	It("polls the deployment on an update and provides rollout status when progressing", func() {
		exampleDepl := getExampleDeployment(namespace)
		d := deployment.NewDeployment(exampleDepl, timeout)
		_, err := d.CreateOrPatch(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())

		th.SimulateDeploymentReplicaReady(th.GetName(exampleDepl))

		// update image of the deployment
		exampleDepl.Spec.Template.Spec.Containers[0].Image = "new-test-container-image"
		d = deployment.NewDeployment(exampleDepl, timeout)
		_, err = d.CreateOrPatch(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())

		// set deployment to progressing
		th.SimulateDeploymentProgressing(th.GetName(exampleDepl))

		// provides the rollout status after the poll timeout exceeded
		_, err = d.CreateOrPatch(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(d.GetRolloutStatus()).NotTo(BeNil())
		Expect(*d.GetRolloutStatus()).To(Equal(deployment.DeploymentPollProgressing))
		Expect(d.GetRolloutMessage()).To(Equal(fmt.Sprintf("%s - 3/4 replicas updated", exampleDepl.Name)))

		// set deployment to succeed
		th.SimulateDeploymentReplicaReady(th.GetName(exampleDepl))

		_, err = d.CreateOrPatch(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(d.GetRolloutStatus()).NotTo(BeNil())
		Expect(*d.GetRolloutStatus()).To(Equal(deployment.DeploymentPollCompleted))
		Expect(d.GetRolloutMessage()).To(Equal(fmt.Sprintf("%s Completed", exampleDepl.Name)))
	})

	It("polls the deployment on an update and provides rollout DeadlineExceeded status when reached", func() {
		exampleDepl := getExampleDeployment(namespace)
		d := deployment.NewDeployment(exampleDepl, timeout)
		_, err := d.CreateOrPatch(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())

		th.SimulateDeploymentReplicaReady(th.GetName(exampleDepl))

		// update image of the deployment
		exampleDepl.Spec.Template.Spec.Containers[0].Image = "new-test-container-image"
		d = deployment.NewDeployment(exampleDepl, timeout)
		_, err = d.CreateOrPatch(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())

		// set deployment to progressing
		th.SimulateDeploymentProgressDeadlineExceeded(th.GetName(exampleDepl))

		// provides the rollout status after the poll timeout exceeded
		_, err = d.CreateOrPatch(ctx, h)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(d.GetRolloutStatus()).NotTo(BeNil())
		Expect(*d.GetRolloutStatus()).To(Equal(deployment.DeploymentPollProgressDeadlineExceeded))
		Expect(d.GetRolloutMessage()).To(ContainSubstring(fmt.Sprintf("%s ProgressDeadlineExceeded", exampleDepl.Name)))
	})

	It("DeleteDeployment deletes existing deployments", func() {
		d, k8sDepl := runDeploymentSuccessfully(namespace)
		// the job exists
		th.GetDeployment(th.GetName(k8sDepl))

		// assert that Delete deletes it properly so the k8sDeployment not found
		// any more
		Expect(d.Delete(ctx, h)).To(Succeed())
		Eventually(func(g Gomega) {
			err := cClient.Get(ctx, th.GetName(k8sDepl), k8sDepl)
			g.Expect(err).To(HaveOccurred())
			var statusErr *k8s_errors.StatusError
			g.Expect(errors.As(err, &statusErr)).To(BeTrue())
			g.Expect(statusErr.Status().Reason).To(Equal(metav1.StatusReasonNotFound))
		}, timeout, interval).Should(Succeed())
	})

})
