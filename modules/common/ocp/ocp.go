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

// Package ocp provides utilities for OpenShift cluster operations and configuration
package ocp

import (
	"context"

	"github.com/openstack-k8s-operators/lib-common/modules/common/helper"

	ocp_config "github.com/openshift/api/config/v1"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/types"
	k8s_utils "k8s.io/utils/net"
)

// IsFipsCluster - Check if OCP has fips enabled which is a day 1 operation
func IsFipsCluster(ctx context.Context, h *helper.Helper) (bool, error) {
	configMap := &corev1.ConfigMap{}
	err := h.GetClient().Get(ctx, types.NamespacedName{Name: "cluster-config-v1", Namespace: "kube-system"}, configMap)
	if err != nil {
		return false, err
	}

	var installConfig map[string]interface{}
	installConfigYAML := configMap.Data["install-config"]
	err = yaml.Unmarshal([]byte(installConfigYAML), &installConfig)
	if err != nil {
		return false, err
	}

	fipsEnabled, ok := installConfig["fips"].(bool)
	if !ok {
		return false, nil
	}
	return fipsEnabled, nil
}

// HasIPv6ClusterNetwork - Check if OCP has an IPv6 cluster network
// Falls back to checking Node PodCIDR and kubernetes service IP families for MicroShift
func HasIPv6ClusterNetwork(ctx context.Context, h *helper.Helper) (bool, error) {
	networkConfig := &ocp_config.Network{}
	err := h.GetClient().Get(ctx, types.NamespacedName{Name: "cluster", Namespace: ""}, networkConfig)
	if err == nil {
		// OCP Network config available
		for _, clusterNetwork := range networkConfig.Status.ClusterNetwork {
			if k8s_utils.IsIPv6CIDRString(clusterNetwork.CIDR) {
				return true, nil
			}
		}
		return false, nil
	}

	// Fallback for MicroShift: check if Network CRD is not available
	if !meta.IsNoMatchError(err) {
		return false, err
	}

	// Check Node PodCIDR
	nodeList := &corev1.NodeList{}
	err = h.GetClient().List(ctx, nodeList)
	if err != nil {
		return false, err
	}

	for _, node := range nodeList.Items {
		if node.Spec.PodCIDR != "" && k8s_utils.IsIPv6CIDRString(node.Spec.PodCIDR) {
			return true, nil
		}
		for _, podCIDR := range node.Spec.PodCIDRs {
			if k8s_utils.IsIPv6CIDRString(podCIDR) {
				return true, nil
			}
		}
	}

	return false, nil
}

// FirstClusterNetworkIsIPv6 - Check if first OCP cluster network is IPv6
// Falls back to checking first Node PodCIDR and kubernetes service IP families for MicroShift
func FirstClusterNetworkIsIPv6(ctx context.Context, h *helper.Helper) (bool, error) {
	networkConfig := &ocp_config.Network{}
	err := h.GetClient().Get(ctx, types.NamespacedName{Name: "cluster", Namespace: ""}, networkConfig)
	if err == nil {
		// OCP Network config available
		for _, clusterNetwork := range networkConfig.Status.ClusterNetwork {
			return k8s_utils.IsIPv6CIDRString(clusterNetwork.CIDR), nil
		}
		return false, nil
	}

	// Fallback for MicroShift: check if Network CRD is not available
	if !meta.IsNoMatchError(err) {
		return false, err
	}

	// Check first Node PodCIDR
	nodeList := &corev1.NodeList{}
	err = h.GetClient().List(ctx, nodeList)
	if err != nil {
		return false, err
	}

	if len(nodeList.Items) > 0 {
		node := nodeList.Items[0]
		if node.Spec.PodCIDR != "" {
			return k8s_utils.IsIPv6CIDRString(node.Spec.PodCIDR), nil
		}
		if len(node.Spec.PodCIDRs) > 0 {
			return k8s_utils.IsIPv6CIDRString(node.Spec.PodCIDRs[0]), nil
		}
	}

	return false, nil
}
