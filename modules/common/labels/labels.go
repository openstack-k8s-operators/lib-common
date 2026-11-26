/*
Copyright 2020 Red Hat

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

// Package labels provides utilities for managing Kubernetes labels and selectors
package labels

import (
	"github.com/openstack-k8s-operators/lib-common/modules/common"
	"github.com/openstack-k8s-operators/lib-common/modules/common/util"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// k8s recommended labels from https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/ .
const (
	K8sAppName      = "app.kubernetes.io/name"
	K8sAppInstance  = "app.kubernetes.io/instance"
	K8sAppVersion   = "app.kubernetes.io/version"
	K8sAppComponent = "app.kubernetes.io/component"
	K8sAppPartOf    = "app.kubernetes.io/part-of"
	K8sAppManagedBy = "app.kubernetes.io/managed-by"
	K8sHostname     = "kubernetes.io/hostname"
)

// GetGroupLabel -
func GetGroupLabel(serviceName string) string {
	return serviceName + ".openstack.org"
}

// GetOwnerUIDLabelSelector -
func GetOwnerUIDLabelSelector(groupLabel string) string {
	return groupLabel + "/uid"
}

// GetOwnerNameSpaceLabelSelector -
func GetOwnerNameSpaceLabelSelector(groupLabel string) string {
	return groupLabel + "/namespace"
}

// GetOwnerNameLabelSelector -
func GetOwnerNameLabelSelector(groupLabel string) string {
	return groupLabel + "/name"
}

// GetLabels - create default labels map, additional custom labels can be passed
func GetLabels(
	obj metav1.Object,
	groupLabel string,
	custom map[string]string,
) map[string]string {
	ownerUIDLabelSelector := GetOwnerUIDLabelSelector(groupLabel)
	ownerNameSpaceLabelSelector := GetOwnerNameSpaceLabelSelector(groupLabel)
	ownerNameLabelSelector := GetOwnerNameLabelSelector(groupLabel)

	// Labels for all objects
	labelSelector := map[string]string{
		ownerUIDLabelSelector:       string(obj.GetUID()),
		ownerNameSpaceLabelSelector: obj.GetNamespace(),
		ownerNameLabelSelector:      obj.GetName(),
	}

	return util.MergeStringMaps(labelSelector, custom)
}

// GetSingleLabelSelector - create a simple label selector for a single key/value label
func GetSingleLabelSelector(
	key string,
	value string,
) metav1.LabelSelector {
	return metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{
				Key:      key,
				Operator: metav1.LabelSelectorOpIn,
				Values: []string{
					value,
				},
			},
		},
	}
}

// GetAppLabelSelector - create a simple label selector for the default AppSelector "service" label
func GetAppLabelSelector(
	name string,
) metav1.LabelSelector {
	return GetSingleLabelSelector(common.AppSelector, name)
}

// GetLabelSelector - utility function that returns a metav1.LabelSelector
// based on the map[string]string that represents the k/v list passed to the
// StatefulSet or Deployment as labelSelector
func GetLabelSelector(
	serviceLabels map[string]string,
) metav1.LabelSelector {
	return metav1.LabelSelector{
		MatchLabels: serviceLabels,
	}
}

// EqualLabelSelectors - returns true if two labelSelectors matches, false
// otherwise
func EqualLabelSelectors(
	l1, l2 metav1.LabelSelector) bool {
	return equality.Semantic.DeepEqual(l1, l2)
}
