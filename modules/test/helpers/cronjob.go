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
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	batchv1 "k8s.io/api/batch/v1"
)

// GetCronJob retrieves a specified CronJob resource from the cluster.
//
// Example usage:
//
//	job := th.GetCronJob(types.NamespacedName{Name: "cell-name", Namespace: "default"})
func (tc *TestHelper) GetCronJob(name types.NamespacedName) *batchv1.CronJob {
	job := &batchv1.CronJob{}
	gomega.Eventually(func(g gomega.Gomega) {
		g.Expect(tc.K8sClient.Get(tc.Ctx, name, job)).Should(gomega.Succeed())
	}, tc.Timeout, tc.Interval).Should(gomega.Succeed())

	return job
}

// AssertCronJobDoesNotExist ensures the CronJob resource does not exist in a k8s cluster.
func (tc *TestHelper) AssertCronJobDoesNotExist(name types.NamespacedName) {
	instance := &batchv1.CronJob{}
	gomega.Eventually(func(g gomega.Gomega) {
		err := tc.K8sClient.Get(tc.Ctx, name, instance)
		g.Expect(k8s_errors.IsNotFound(err)).To(gomega.BeTrue())
	}, tc.Timeout, tc.Interval).Should(gomega.Succeed())
}
