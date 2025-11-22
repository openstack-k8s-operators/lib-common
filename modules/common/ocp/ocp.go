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

	var installConfig map[string]any
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
func HasIPv6ClusterNetwork(ctx context.Context, h *helper.Helper) (bool, error) {
	networkConfig := &ocp_config.Network{}
	err := h.GetClient().Get(ctx, types.NamespacedName{Name: "cluster", Namespace: ""}, networkConfig)
	if err != nil {
		return false, err
	}

	for _, clusterNetwork := range networkConfig.Status.ClusterNetwork {
		if k8s_utils.IsIPv6CIDRString(clusterNetwork.CIDR) {
			return true, nil
		}
	}
	return false, nil
}

// FirstClusterNetworkIsIPv6 - Check if first OCP cluster network is IPv6
func FirstClusterNetworkIsIPv6(ctx context.Context, h *helper.Helper) (bool, error) {
	networkConfig := &ocp_config.Network{}
	err := h.GetClient().Get(ctx, types.NamespacedName{Name: "cluster", Namespace: ""}, networkConfig)
	if err != nil {
		return false, err
	}

	for _, clusterNetwork := range networkConfig.Status.ClusterNetwork {
		return k8s_utils.IsIPv6CIDRString(clusterNetwork.CIDR), nil
	}
	return false, nil
}
