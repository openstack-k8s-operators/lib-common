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

package configmap

import (
	"context"
	"fmt"
	"time"

	"github.com/openstack-k8s-operators/lib-common/modules/common/env"
	"github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	"github.com/openstack-k8s-operators/lib-common/modules/common/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	corev1 "k8s.io/api/core/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
)

// Hash function creates a hash of a ConfigMap's Data and BinaryData fields and
// returns it as a safe encoded string.
func Hash(configMap *corev1.ConfigMap) (string, error) {
	type ConfigMapData struct {
		Data       map[string]string `json:"data,omitempty" protobuf:"bytes,2,rep,name=data"`
		BinaryData map[string][]byte `json:"binaryData,omitempty" protobuf:"bytes,3,rep,name=binaryData"`
	}

	if configMap == nil {
		return "", fmt.Errorf("nil ConfigMap doesn't have data to hash")
	}

	data := ConfigMapData{
		Data:       configMap.Data,
		BinaryData: configMap.BinaryData,
	}
	return util.ObjectHash(data)
}

// createOrPatchConfigMap -
func createOrPatchConfigMap(
	ctx context.Context,
	h *helper.Helper,
	obj client.Object,
	cm util.Template,
) (string, controllerutil.OperationResult, error) {
	data := make(map[string]string)

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:        cm.Name,
			Namespace:   cm.Namespace,
			Annotations: cm.Annotations,
		},
		Data: data,
	}

	// create or update the CM
	op, err := controllerutil.CreateOrPatch(ctx, h.GetClient(), configMap, func() error {

		configMap.Labels = cm.Labels
		// add data from templates
		renderedTemplateData, err := util.GetTemplateData(cm)
		if err != nil {
			return err
		}
		configMap.Data = renderedTemplateData
		// add provided custom data to configMap.Data
		// Note: this can overwrite data rendered from GetTemplateData() if key is same
		if len(cm.CustomData) > 0 {
			for k, v := range cm.CustomData {
				configMap.Data[k] = v
			}
		}

		if !cm.SkipSetOwner {
			err := controllerutil.SetControllerReference(obj, configMap, h.GetScheme())
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return "", op, fmt.Errorf("error create/updating configmap: %v", err)
	}

	configMapHash, err := Hash(configMap)
	if err != nil {
		return "", op, fmt.Errorf("error calculating configuration hash: %v", err)
	}

	return configMapHash, op, nil
}

// createOrGetCustomConfigMap -
func createOrGetCustomConfigMap(
	ctx context.Context,
	h *helper.Helper,
	obj client.Object,
	cm util.Template,
) (string, error) {
	// Check if this configMap already exists
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:        cm.Name,
			Namespace:   cm.Namespace,
			Labels:      cm.Labels,
			Annotations: cm.Annotations,
		},
		Data: map[string]string{},
	}
	foundConfigMap := &corev1.ConfigMap{}
	err := h.GetClient().Get(ctx, types.NamespacedName{Name: cm.Name, Namespace: cm.Namespace}, foundConfigMap)
	if err != nil && k8s_errors.IsNotFound(err) {
		if !cm.SkipSetOwner {
			err := controllerutil.SetControllerReference(obj, configMap, h.GetScheme())
			if err != nil {
				return "", err
			}
		}

		h.GetLogger().Info(fmt.Sprintf("Creating a new ConfigMap %s in namespace %s", cm.Namespace, cm.Name))
		err = h.GetClient().Create(ctx, configMap)
		if err != nil {
			return "", err
		}
	} else {
		// use data from already existing custom configmap
		configMap.Data = foundConfigMap.Data
	}

	configMapHash, err := Hash(configMap)
	if err != nil {
		return "", fmt.Errorf("error calculating configuration hash: %v", err)
	}

	return configMapHash, nil
}

// EnsureConfigMaps - get all configmaps required, verify they exist and add the hash to env and status
func EnsureConfigMaps(
	ctx context.Context,
	h *helper.Helper,
	obj client.Object,
	cms []util.Template,
	envVars *map[string]env.Setter,
) error {
	var err error

	for _, cm := range cms {
		var hash string
		var op controllerutil.OperationResult

		if cm.Type != util.TemplateTypeCustom {
			hash, op, err = createOrPatchConfigMap(ctx, h, obj, cm)
		} else {
			hash, err = createOrGetCustomConfigMap(ctx, h, obj, cm)
			// set op to OperationResultNone because createOrGetCustomConfigMap does not return an op
			// and it will add log entries bellow with none operation
			op = controllerutil.OperationResult(controllerutil.OperationResultNone)
		}
		if err != nil {
			return err
		}
		if op != controllerutil.OperationResultNone {
			h.GetLogger().Info(fmt.Sprintf("ConfigMap %s successfully reconciled - operation: %s", cm.Name, string(op)))
		}
		if envVars != nil {
			(*envVars)[cm.Name] = env.SetValue(hash)
		}
	}

	return nil
}

// GetConfigMaps - get all configmaps required, verify they exist and add the hash to env and status
func GetConfigMaps(
	ctx context.Context,
	h *helper.Helper,
	obj client.Object,
	configMaps []string,
	namespace string,
	envVars *map[string]env.Setter,
) ([]util.Hash, error) {
	hashes := []util.Hash{}

	for _, cm := range configMaps {
		_, hash, err := GetConfigMapAndHashWithName(ctx, h, cm, namespace)
		if err != nil {
			return nil, err
		}
		(*envVars)[cm] = env.SetValue(hash)
		hashes = append(hashes, util.Hash{Name: cm, Hash: hash})
	}

	return hashes, nil
}

// GetConfigMapAndHashWithName -
func GetConfigMapAndHashWithName(
	ctx context.Context,
	h *helper.Helper,
	configMapName string,
	namespace string,
) (*corev1.ConfigMap, string, error) {

	configMap := &corev1.ConfigMap{}
	err := h.GetClient().Get(ctx, types.NamespacedName{Name: configMapName, Namespace: namespace}, configMap)
	if err != nil && k8s_errors.IsNotFound(err) {
		h.GetLogger().Error(err, configMapName+" ConfigMap not found!", "Instance.Namespace", namespace, "ConfigMap.Name", configMapName)
		return configMap, "", err
	}
	configMapHash, err := Hash(configMap)
	if err != nil {
		return configMap, "", fmt.Errorf("error calculating configuration hash: %v", err)
	}
	return configMap, configMapHash, nil
}

// GetConfigMap - Get config map
//
// if the config map is not found, requeue after requeueTimeout
func GetConfigMap(
	ctx context.Context,
	h *helper.Helper,
	object client.Object,
	configMapName string,
	requeueTimeout time.Duration,
) (*corev1.ConfigMap, ctrl.Result, error) {

	configMap := &corev1.ConfigMap{}
	err := h.GetClient().Get(ctx, types.NamespacedName{Name: configMapName, Namespace: object.GetNamespace()}, configMap)
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			msg := fmt.Sprintf("%s config map does not exist: %v", configMapName, err)
			util.LogForObject(h, msg, object)

			return configMap, ctrl.Result{RequeueAfter: requeueTimeout}, nil
		}
		msg := fmt.Sprintf("Error getting %s config map: %v", configMapName, err)
		err = util.WrapErrorForObject(msg, object, err)

		return configMap, ctrl.Result{}, err
	}

	return configMap, ctrl.Result{}, nil
}

// GetConfigMapListWithLabel - Get all configmaps in namespace of the obj matching label selector
func GetConfigMapListWithLabel(
	ctx context.Context,
	h *helper.Helper,
	namespace string,
	labelSelectorMap map[string]string,
) (*corev1.ConfigMapList, error) {
	configMaps := &corev1.ConfigMapList{}
	listOpts := []client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabels(labelSelectorMap),
	}

	if err := h.GetClient().List(ctx, configMaps, listOpts...); err != nil {
		err = fmt.Errorf("Error listing config map for labels: %v - %v", labelSelectorMap, err)
		return nil, err
	}

	return configMaps, nil
}

// EnsureOperatorConfigMap - Create ConfigMap consumed by operators
func EnsureOperatorConfigMap(
	ctx context.Context,
	h *helper.Helper,
	obj client.Object,
	name string,
	data map[string]string,
	labels map[string]string,
	scope string,
) error {
	cmLabels := util.MergeStringMaps(
		labels,
		map[string]string{
			"operators.openstack.org/consumer": "operator",
			"operators.openstack.org/scope":    scope,
		},
	)

	cms := []util.Template{
		{
			Name:       name,
			Namespace:  obj.GetNamespace(),
			Type:       util.TemplateTypeConfig,
			CustomData: data,
			Labels:     cmLabels,
		},
	}

	envVars := make(map[string]env.Setter)
	return EnsureConfigMaps(ctx, h, obj, cms, &envVars)
}

// GetOperatorConfigMapList - Get ConfigMap for operators, tagged for the specific scope
func GetOperatorConfigMapList(
	ctx context.Context,
	h *helper.Helper,
	namespace string,
	scope string,
) (*corev1.ConfigMapList, *corev1.ConfigMapList, error) {

	globalLabelSelectorMap := map[string]string{
		"operators.openstack.org/consumer": "operator",
		"operators.openstack.org/scope":    "global",
	}
	globalList, err := GetConfigMapListWithLabel(ctx, h, namespace, globalLabelSelectorMap)
	if err != nil {
		return nil, nil, err
	}

	localLabelSelectorMap := map[string]string{
		"operators.openstack.org/consumer": "operator",
		"operators.openstack.org/scope":    scope,
	}
	localList, err := GetConfigMapListWithLabel(ctx, h, namespace, localLabelSelectorMap)
	if err != nil {
		return nil, nil, err
	}

	return globalList, localList, err
}

// MergeConfigMapData - Merge multiple ConfigMaps and generate a single map
func MergeConfigMapData(
	list *corev1.ConfigMapList,
) map[string]string {
	result := map[string]string{}

	for _, cm := range list.Items {
		result = util.MergeStringMaps(result, cm.Data)
	}

	return result
}
