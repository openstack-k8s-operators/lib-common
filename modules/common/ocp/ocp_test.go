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

package ocp

import (
	"context"
	"testing"

	"github.com/openstack-k8s-operators/lib-common/modules/common/helper"

	. "github.com/onsi/gomega"
	ocp_config "github.com/openshift/api/config/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func setupHelper(objs ...client.Object) (*helper.Helper, error) {
	s := scheme.Scheme
	err := ocp_config.AddToScheme(s)
	if err != nil {
		return nil, err
	}

	fakeClient := fake.NewClientBuilder().
		WithScheme(s).
		WithObjects(objs...).
		Build()

	// Create a minimal namespace object for helper
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-namespace",
		},
	}

	return helper.NewHelper(ns, fakeClient, nil, s, ctrl.Log)
}

func TestHasIPv6ClusterNetwork_OCP_IPv4(t *testing.T) {
	g := NewWithT(t)

	networkConfig := &ocp_config.Network{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster",
		},
		Status: ocp_config.NetworkStatus{
			ClusterNetwork: []ocp_config.ClusterNetworkEntry{
				{CIDR: "10.128.0.0/14"},
				{CIDR: "10.132.0.0/14"},
			},
		},
	}

	h, err := setupHelper(networkConfig)
	g.Expect(err).NotTo(HaveOccurred())

	hasIPv6, err := HasIPv6ClusterNetwork(context.TODO(), h)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(hasIPv6).To(BeFalse())
}

func TestHasIPv6ClusterNetwork_OCP_IPv6(t *testing.T) {
	g := NewWithT(t)

	networkConfig := &ocp_config.Network{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster",
		},
		Status: ocp_config.NetworkStatus{
			ClusterNetwork: []ocp_config.ClusterNetworkEntry{
				{CIDR: "fd01::/48"},
				{CIDR: "fd02::/48"},
			},
		},
	}

	h, err := setupHelper(networkConfig)
	g.Expect(err).NotTo(HaveOccurred())

	hasIPv6, err := HasIPv6ClusterNetwork(context.TODO(), h)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(hasIPv6).To(BeTrue())
}

func TestHasIPv6ClusterNetwork_OCP_DualStack(t *testing.T) {
	g := NewWithT(t)

	networkConfig := &ocp_config.Network{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster",
		},
		Status: ocp_config.NetworkStatus{
			ClusterNetwork: []ocp_config.ClusterNetworkEntry{
				{CIDR: "10.128.0.0/14"},
				{CIDR: "fd01::/48"},
			},
		},
	}

	h, err := setupHelper(networkConfig)
	g.Expect(err).NotTo(HaveOccurred())

	hasIPv6, err := HasIPv6ClusterNetwork(context.TODO(), h)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(hasIPv6).To(BeTrue())
}

func TestHasIPv6ClusterNetwork_MicroShift_IPv4_NodePodCIDR(t *testing.T) {
	g := NewWithT(t)

	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "microshift-0",
		},
		Spec: corev1.NodeSpec{
			PodCIDR: "10.42.0.0/24",
		},
	}

	h, err := setupHelper(node)
	g.Expect(err).NotTo(HaveOccurred())

	hasIPv6, err := HasIPv6ClusterNetwork(context.TODO(), h)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(hasIPv6).To(BeFalse())
}

func TestHasIPv6ClusterNetwork_MicroShift_IPv6_NodePodCIDR(t *testing.T) {
	g := NewWithT(t)

	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "microshift-0",
		},
		Spec: corev1.NodeSpec{
			PodCIDR: "fd01::/48",
		},
	}

	h, err := setupHelper(node)
	g.Expect(err).NotTo(HaveOccurred())

	hasIPv6, err := HasIPv6ClusterNetwork(context.TODO(), h)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(hasIPv6).To(BeTrue())
}

func TestHasIPv6ClusterNetwork_MicroShift_DualStack_NodePodCIDRs(t *testing.T) {
	g := NewWithT(t)

	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "microshift-0",
		},
		Spec: corev1.NodeSpec{
			PodCIDRs: []string{"10.42.0.0/24", "fd01::/48"},
		},
	}

	h, err := setupHelper(node)
	g.Expect(err).NotTo(HaveOccurred())

	hasIPv6, err := HasIPv6ClusterNetwork(context.TODO(), h)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(hasIPv6).To(BeTrue())
}

func TestHasIPv6ClusterNetwork_MicroShift_NoPodCIDR(t *testing.T) {
	g := NewWithT(t)

	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "microshift-0",
		},
		Spec: corev1.NodeSpec{},
	}

	h, err := setupHelper(node)
	g.Expect(err).NotTo(HaveOccurred())

	hasIPv6, err := HasIPv6ClusterNetwork(context.TODO(), h)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(hasIPv6).To(BeFalse())
}

func TestFirstClusterNetworkIsIPv6_OCP_IPv4First(t *testing.T) {
	g := NewWithT(t)

	networkConfig := &ocp_config.Network{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster",
		},
		Status: ocp_config.NetworkStatus{
			ClusterNetwork: []ocp_config.ClusterNetworkEntry{
				{CIDR: "10.128.0.0/14"},
				{CIDR: "fd01::/48"},
			},
		},
	}

	h, err := setupHelper(networkConfig)
	g.Expect(err).NotTo(HaveOccurred())

	isIPv6, err := FirstClusterNetworkIsIPv6(context.TODO(), h)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(isIPv6).To(BeFalse())
}

func TestFirstClusterNetworkIsIPv6_OCP_IPv6First(t *testing.T) {
	g := NewWithT(t)

	networkConfig := &ocp_config.Network{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster",
		},
		Status: ocp_config.NetworkStatus{
			ClusterNetwork: []ocp_config.ClusterNetworkEntry{
				{CIDR: "fd01::/48"},
				{CIDR: "10.128.0.0/14"},
			},
		},
	}

	h, err := setupHelper(networkConfig)
	g.Expect(err).NotTo(HaveOccurred())

	isIPv6, err := FirstClusterNetworkIsIPv6(context.TODO(), h)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(isIPv6).To(BeTrue())
}

func TestFirstClusterNetworkIsIPv6_MicroShift_IPv4First(t *testing.T) {
	g := NewWithT(t)

	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "microshift-0",
		},
		Spec: corev1.NodeSpec{
			PodCIDRs: []string{"10.42.0.0/24", "fd01::/48"},
		},
	}

	h, err := setupHelper(node)
	g.Expect(err).NotTo(HaveOccurred())

	isIPv6, err := FirstClusterNetworkIsIPv6(context.TODO(), h)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(isIPv6).To(BeFalse())
}

func TestFirstClusterNetworkIsIPv6_MicroShift_IPv6First(t *testing.T) {
	g := NewWithT(t)

	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "microshift-0",
		},
		Spec: corev1.NodeSpec{
			PodCIDRs: []string{"fd01::/48", "10.42.0.0/24"},
		},
	}

	h, err := setupHelper(node)
	g.Expect(err).NotTo(HaveOccurred())

	isIPv6, err := FirstClusterNetworkIsIPv6(context.TODO(), h)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(isIPv6).To(BeTrue())
}

func TestFirstClusterNetworkIsIPv6_MicroShift_NoPodCIDR(t *testing.T) {
	g := NewWithT(t)

	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "microshift-0",
		},
		Spec: corev1.NodeSpec{},
	}

	h, err := setupHelper(node)
	g.Expect(err).NotTo(HaveOccurred())

	isIPv6, err := FirstClusterNetworkIsIPv6(context.TODO(), h)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(isIPv6).To(BeFalse())
}

func TestIsFipsCluster(t *testing.T) {
	g := NewWithT(t)

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster-config-v1",
			Namespace: "kube-system",
		},
		Data: map[string]string{
			"install-config": "apiVersion: v1\nbaseDomain: example.com\nfips: true",
		},
	}

	h, err := setupHelper(configMap)
	g.Expect(err).NotTo(HaveOccurred())

	isFips, err := IsFipsCluster(context.TODO(), h)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(isFips).To(BeTrue())
}

func TestIsFipsCluster_NotEnabled(t *testing.T) {
	g := NewWithT(t)

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster-config-v1",
			Namespace: "kube-system",
		},
		Data: map[string]string{
			"install-config": "apiVersion: v1\nbaseDomain: example.com\nfips: false",
		},
	}

	h, err := setupHelper(configMap)
	g.Expect(err).NotTo(HaveOccurred())

	isFips, err := IsFipsCluster(context.TODO(), h)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(isFips).To(BeFalse())
}

func TestIsFipsCluster_NotSet(t *testing.T) {
	g := NewWithT(t)

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster-config-v1",
			Namespace: "kube-system",
		},
		Data: map[string]string{
			"install-config": "apiVersion: v1\nbaseDomain: example.com",
		},
	}

	h, err := setupHelper(configMap)
	g.Expect(err).NotTo(HaveOccurred())

	isFips, err := IsFipsCluster(context.TODO(), h)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(isFips).To(BeFalse())
}
