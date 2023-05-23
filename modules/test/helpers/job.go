/*
Copyright 2022 Red Hat
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

package helpers

import (
	"github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	batchv1 "k8s.io/api/batch/v1"
)

// GetJob -
func (tc *TestHelper) GetJob(name types.NamespacedName) *batchv1.Job {
	job := &batchv1.Job{}
	gomega.Eventually(func(g gomega.Gomega) {
		g.Expect(tc.K8sClient.Get(tc.Ctx, name, job)).Should(gomega.Succeed())
	}, tc.Timeout, tc.Interval).Should(gomega.Succeed())

	return job
}

// ListJobs -
func (tc *TestHelper) ListJobs(namespace string) *batchv1.JobList {
	jobs := &batchv1.JobList{}
	gomega.Expect(tc.K8sClient.List(tc.Ctx, jobs, client.InNamespace(namespace))).Should(gomega.Succeed())

	return jobs
}

// SimulateJobFailure -
func (tc *TestHelper) SimulateJobFailure(name types.NamespacedName) {
	gomega.Eventually(func(g gomega.Gomega) {
		job := tc.GetJob(name)

		// Simulate that the job is failed
		job.Status.Failed = 1
		job.Status.Active = 0
		g.Expect(tc.K8sClient.Status().Update(tc.Ctx, job)).To(gomega.Succeed())

	}, tc.Timeout, tc.Interval).Should(gomega.Succeed())
	tc.Logger.Info("Simulated Job failure", "on", name)
}

// SimulateJobSuccess -
func (tc *TestHelper) SimulateJobSuccess(name types.NamespacedName) {
	gomega.Eventually(func(g gomega.Gomega) {
		job := tc.GetJob(name)

		// Simulate that the job is succeeded
		job.Status.Succeeded = 1
		job.Status.Active = 0
		g.Expect(tc.K8sClient.Status().Update(tc.Ctx, job)).To(gomega.Succeed())
	}, tc.Timeout, tc.Interval).Should(gomega.Succeed())

	tc.Logger.Info("Simulated Job success", "on", name)
}
