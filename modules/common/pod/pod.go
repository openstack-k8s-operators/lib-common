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

package pod

import (
	"context"
	"fmt"

	"github.com/openstack-k8s-operators/lib-common/modules/common/helper"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

// GetPodListWithLabel - Get all pods in namespace of the obj matching label selector
func GetPodListWithLabel(
	ctx context.Context,
	h *helper.Helper,
	namespace string,
	labelSelectorMap map[string]string,
) (*corev1.PodList, error) {

	labelSelectorString := labels.Set(labelSelectorMap).String()

	// use kclient to not use a cached client to be able to list pods in namespace which are not cached
	// otherwise we hit "Error listing pods for labels: map[ ... ] - unable to get: default because of unknown namespace for the cache"
	podList, err := h.GetKClient().CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: labelSelectorString})
	if err != nil {
		err = fmt.Errorf("Error listing pods for labels: %v - %w", labelSelectorMap, err)
		return nil, err
	}

	return podList, nil
}
