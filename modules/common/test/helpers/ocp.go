/*
Copyright 2024 Red Hat
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
	"sigs.k8s.io/controller-runtime/pkg/client"

	ocp_config "github.com/openshift/api/config/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateClusterNetworkConfig creates a fake cluster network config CR
func (tc *TestHelper) CreateClusterNetworkConfig() client.Object {
	instance := &ocp_config.Network{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster",
			Namespace: "",
		},
		Status: ocp_config.NetworkStatus{
			ClusterNetwork: []ocp_config.ClusterNetworkEntry{
				{
					CIDR:       "172.16.0.0/25",
					HostPrefix: 24,
				},
			},
		},
	}
	gomega.Expect(tc.K8sClient.Create(tc.Ctx, instance)).Should(gomega.Succeed())

	return instance
}
