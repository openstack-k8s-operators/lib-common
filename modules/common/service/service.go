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

package service

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	"github.com/openstack-k8s-operators/lib-common/modules/common/util"
	ctrl "sigs.k8s.io/controller-runtime"
)

// NewService returns an initialized Service.
func NewService(
	service *corev1.Service,
	labels map[string]string,
	timeout int,
) *Service {
	return &Service{
		service: service,
		timeout: timeout,
	}
}

// GenericService func
func GenericService(svcInfo *GenericServiceDetails) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      svcInfo.Name,
			Namespace: svcInfo.Namespace,
			Labels:    svcInfo.Labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: svcInfo.Selector,
			Ports: []corev1.ServicePort{
				{
					Name: svcInfo.Port.Name,
					Port: svcInfo.Port.Port,
					// corev1.ProtocolTCP/ corev1.ProtocolUDP/ corev1.ProtocolSCTP
					// - https://pkg.go.dev/k8s.io/api@v0.23.6/core/v1#Protocol
					Protocol: svcInfo.Port.Protocol,
				},
			},
			ClusterIP: svcInfo.ClusterIP,
		},
	}
}

// CreateOrPatch - creates or patches a service, reconciles after Xs if object won't exist.
func (s *Service) CreateOrPatch(
	ctx context.Context,
	h *helper.Helper,
) (ctrl.Result, error) {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.service.Name,
			Namespace: s.service.Namespace,
		},
	}

	op, err := controllerutil.CreateOrPatch(ctx, h.GetClient(), service, func() error {
		service.Labels = util.MergeStringMaps(service.Labels, s.service.Labels)
		service.Annotations = s.service.Annotations
		service.Spec = s.service.Spec

		err := controllerutil.SetControllerReference(h.GetBeforeObject(), service, h.GetScheme())
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			h.GetLogger().Info(fmt.Sprintf("Service %s not found, reconcile in %ds", service.Name, s.timeout))
			return ctrl.Result{RequeueAfter: time.Duration(s.timeout) * time.Second}, nil
		}
		return ctrl.Result{}, err
	}
	if op != controllerutil.OperationResultNone {
		h.GetLogger().Info(fmt.Sprintf("Service %s - %s", service.Name, op))
	}

	return ctrl.Result{}, nil
}

// Delete - delete a service.
func (s *Service) Delete(
	ctx context.Context,
	h *helper.Helper,
) error {

	err := h.GetClient().Delete(ctx, s.service)
	if err != nil && !k8s_errors.IsNotFound(err) {
		err = fmt.Errorf("Error deleting service %s: %v", s.service.Name, err)
		return err
	}

	return nil
}

// DeleteServicesWithLabel - Delete all services in namespace of the obj matching label selector
func DeleteServicesWithLabel(
	ctx context.Context,
	h *helper.Helper,
	obj metav1.Object,
	labelSelectorMap map[string]string,
) error {
	// Service have not implemented DeleteAllOf
	// https://github.com/operator-framework/operator-sdk/issues/3101
	// https://github.com/kubernetes/kubernetes/issues/68468#issuecomment-419981870
	// delete services
	serviceList := &corev1.ServiceList{}
	listOpts := []client.ListOption{
		client.InNamespace(obj.GetNamespace()),
		client.MatchingLabels(labelSelectorMap),
	}

	if err := h.GetClient().List(ctx, serviceList, listOpts...); err != nil {
		err = fmt.Errorf("Error listing services for %s: %v", obj.GetName(), err)
		return err
	}

	// delete all pods
	for _, pod := range serviceList.Items {
		err := h.GetClient().Delete(ctx, &pod)
		if err != nil && !k8s_errors.IsNotFound(err) {
			err = fmt.Errorf("Error deleting service %s: %v", pod.Name, err)
			return err
		}
	}

	return nil
}

// GetServicesListWithLabel - Get all services in namespace of the obj matching label selector
func GetServicesListWithLabel(
	ctx context.Context,
	h *helper.Helper,
	namespace string,
	labelSelectorMap map[string]string,
) (*corev1.ServiceList, error) {

	labelSelectorString := labels.Set(labelSelectorMap).String()

	// use kclient to not use a cached client to be able to list services in namespace which are not cached
	// otherwise we hit "Error listing services for labels: map[ ... ] - unable to get: default because of unknown namespace for the cache"
	serviceList, err := h.GetKClient().CoreV1().Services(namespace).List(ctx, metav1.ListOptions{LabelSelector: labelSelectorString})
	if err != nil {
		err = fmt.Errorf("Error listing services for labels: %v - %v", labelSelectorMap, err)
		return nil, err
	}

	return serviceList, nil
}
