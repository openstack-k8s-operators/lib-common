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
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	corev1 "k8s.io/api/core/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	"github.com/openstack-k8s-operators/lib-common/modules/common/util"
	ctrl "sigs.k8s.io/controller-runtime"
)

// NewService returns an initialized Service.
func NewService(
	service *corev1.Service,
	timeout time.Duration,
	override *OverrideSpec,
) (*Service, error) {
	svc := &Service{
		service:         service,
		serviceHostname: fmt.Sprintf("%s.%s.svc", service.Name, service.GetNamespace()),
		timeout:         timeout,
	}

	// patch service with possible overrides of Labels, Annotations and Spec
	if override != nil {
		if override.EmbeddedLabelsAnnotations != nil {
			if override.Labels != nil {
				svc.service.Labels = util.MergeStringMaps(override.Labels, service.Labels)
			}
			if override.Annotations != nil {
				svc.service.Annotations = util.MergeStringMaps(override.Annotations, service.Annotations)
			}
		}
		if override.Spec != nil {
			originalSpec, err := json.Marshal(service.Spec)
			if err != nil {
				return svc, fmt.Errorf("error marshalling Service Spec: %w", err)
			}

			patch, err := json.Marshal(override.Spec)
			if err != nil {
				return svc, fmt.Errorf("error marshalling Service Spec override: %w", err)
			}

			patchedJSON, err := strategicpatch.StrategicMergePatch(originalSpec, patch, corev1.ServiceSpec{})
			if err != nil {
				return svc, fmt.Errorf("error patching Service Spec: %w", err)
			}

			patchedSpec := corev1.ServiceSpec{}
			err = json.Unmarshal(patchedJSON, &patchedSpec)
			if err != nil {
				return svc, fmt.Errorf("error unmarshalling patched Service Spec: %w", err)
			}
			svc.service.Spec = patchedSpec
		}
	}

	return svc, nil
}

// GetClusterIPs - returns the cluster IPs of the created service
func (s *Service) GetClusterIPs() []string {
	return s.clusterIPs
}

// GetIPFamilies - returns the IP families of the created service
func (s *Service) GetIPFamilies() []corev1.IPFamily {
	return s.ipFamilies
}

// GetExternalIPs - returns a list of external IPs of the created service
func (s *Service) GetExternalIPs() []string {
	return s.externalIPs
}

// GetServiceHostname - returns the service hostname
func (s *Service) GetServiceHostname() string {
	return s.serviceHostname
}

// GetServiceHostnamePort - returns the service hostname with port if service port
// is not nil, otherwise returns GetServiceHostname()
func (s *Service) GetServiceHostnamePort() (string, string) {
	servicePort := GetServicesPortDetails(s.service, s.service.Name)
	if servicePort != nil {
		return s.GetServiceHostname(),
			strconv.FormatInt(int64(servicePort.Port), 10)
	}

	return s.GetServiceHostname(), ""
}

// GetLabels - returns labels of the service
func (s *Service) GetLabels() map[string]string {
	return s.service.Labels
}

// GetAnnotations - returns annotations of the service
func (s *Service) GetAnnotations() map[string]string {
	return s.service.Annotations
}

// GetSpec - returns the spec of the service
func (s *Service) GetSpec() *corev1.ServiceSpec {
	spec := s.service.Spec
	return &spec
}

// GetServiceType - returns type of the service spec
func (s *Service) GetServiceType() corev1.ServiceType {
	return s.service.Spec.Type
}

// AddAnnotation - Adds annotation and merges it with the current set
func (s *Service) AddAnnotation(anno map[string]string) {
	s.service.Annotations = util.MergeStringMaps(s.service.Annotations, anno)
}

// AddAnnotation - Adds annotation and merges it with the current set
func (s *RoutedOverrideSpec) AddAnnotation(anno map[string]string) {
	if s.EmbeddedLabelsAnnotations == nil {
		s.EmbeddedLabelsAnnotations = &EmbeddedLabelsAnnotations{}
	}
	s.Annotations = util.MergeMaps(s.Annotations, anno)
}

// AddLabel - Adds label and merges it with the current set
func (s *RoutedOverrideSpec) AddLabel(label map[string]string) {
	if s.EmbeddedLabelsAnnotations == nil {
		s.EmbeddedLabelsAnnotations = &EmbeddedLabelsAnnotations{}
	}

	s.Labels = util.MergeMaps(s.Labels, label)
}

// AddAnnotation - Adds annotation and merges it with the current set
func (s *OverrideSpec) AddAnnotation(anno map[string]string) {
	if s.EmbeddedLabelsAnnotations == nil {
		s.EmbeddedLabelsAnnotations = &EmbeddedLabelsAnnotations{}
	}
	s.Annotations = util.MergeMaps(s.Annotations, anno)
}

// AddLabel - Adds label and merges it with the current set
func (s *OverrideSpec) AddLabel(label map[string]string) {
	if s.EmbeddedLabelsAnnotations == nil {
		s.EmbeddedLabelsAnnotations = &EmbeddedLabelsAnnotations{}
	}
	s.Labels = util.MergeMaps(s.Labels, label)
}

// GetAPIEndpoint - returns the API endpoint URL for the service to register in keystone.
func (s *Service) GetAPIEndpoint(endpointURL *string, protocol *Protocol, path string) (string, error) {
	var apiEndpoint *url.URL
	var err error
	if endpointURL != nil {
		apiEndpoint, err = url.Parse(*endpointURL)
		if err != nil {
			return "", err
		}
	} else {
		hostname, port := s.GetServiceHostnamePort()

		var endptURL string
		if protocol != nil &&
			((*protocol == ProtocolHTTP && port == "80") ||
				(*protocol == ProtocolHTTPS && port == "443")) {
			endptURL = fmt.Sprintf("%s%s", EndptProtocol(protocol), hostname)
		} else {
			endptURL = fmt.Sprintf("%s%s:%s", EndptProtocol(protocol), hostname, port)
		}

		// Do not include the path in parsing check because %(project_id)s
		// is invalid without being encoded, but they should not be encoded in the actual endpoint
		apiEndpoint, err = url.Parse(endptURL)
		if err != nil {
			return "", err
		}
	}

	return apiEndpoint.String() + path, nil
}

// ToOverrideServiceSpec - convert corev1.ServiceSpec to OverrideServiceSpec
func (s *Service) ToOverrideServiceSpec() (*OverrideServiceSpec, error) {
	overrideServiceSpec := &OverrideServiceSpec{}

	serviceSpec := s.GetSpec()
	if serviceSpec != nil {
		serviceSpecBytes, err := json.Marshal(serviceSpec)
		if err != nil {
			return nil, fmt.Errorf("error marshalling Service Spec: %w", err)
		}

		err = json.Unmarshal(serviceSpecBytes, overrideServiceSpec)
		if err != nil {
			return nil, fmt.Errorf("error unmarshalling service OverrideSpec: %w", err)
		}
	}

	return overrideServiceSpec, nil
}

// GenericService func
func GenericService(svcInfo *GenericServiceDetails) *corev1.Service {
	ports := svcInfo.Ports
	if len(svcInfo.Ports) == 0 {
		ports = []corev1.ServicePort{
			{
				Name: svcInfo.Port.Name,
				Port: svcInfo.Port.Port,
				// corev1.ProtocolTCP/ corev1.ProtocolUDP/ corev1.ProtocolSCTP
				// - https://pkg.go.dev/k8s.io/api@v0.23.6/core/v1#Protocol
				Protocol: svcInfo.Port.Protocol,
			},
		}

	}
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      svcInfo.Name,
			Namespace: svcInfo.Namespace,
			Labels:    svcInfo.Labels,
		},
		Spec: corev1.ServiceSpec{
			Selector:                 svcInfo.Selector,
			Ports:                    ports,
			ClusterIP:                svcInfo.ClusterIP,
			Type:                     corev1.ServiceTypeClusterIP,
			PublishNotReadyAddresses: svcInfo.PublishNotReadyAddresses,
		},
	}
}

// MetalLBService func
// NOTE: (mschuppert) deprecated, can be removed when external endpoint creation moved to openstack-operator
func MetalLBService(svcInfo *MetalLBServiceDetails) *corev1.Service {
	ports := svcInfo.Ports
	if len(svcInfo.Ports) == 0 {
		ports = []corev1.ServicePort{
			{
				Name: svcInfo.Port.Name,
				Port: svcInfo.Port.Port,
				// corev1.ProtocolTCP/ corev1.ProtocolUDP/ corev1.ProtocolSCTP
				// - https://pkg.go.dev/k8s.io/api@v0.23.6/core/v1#Protocol
				Protocol: svcInfo.Port.Protocol,
			},
		}

	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        svcInfo.Name,
			Namespace:   svcInfo.Namespace,
			Annotations: svcInfo.Annotations,
			Labels:      svcInfo.Labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: svcInfo.Selector,
			Ports:    ports,
			Type:     corev1.ServiceTypeLoadBalancer,
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
		service.Labels = util.MergeStringMaps(s.service.Labels, service.Labels)
		service.Annotations = util.MergeStringMaps(s.service.Annotations, service.Annotations)
		service.Spec = s.service.Spec

		err := controllerutil.SetControllerReference(h.GetBeforeObject(), service, h.GetScheme())
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			h.GetLogger().Info(fmt.Sprintf("Service %s not found, reconcile in %s", service.Name, s.timeout))
			return ctrl.Result{RequeueAfter: s.timeout}, nil
		}
		return ctrl.Result{}, err
	}
	if op != controllerutil.OperationResultNone {
		h.GetLogger().Info(fmt.Sprintf("Service %s - %s", service.Name, op))
	}

	// update the service instance with the ip/host information
	s.clusterIPs = service.Spec.ClusterIPs
	s.ipFamilies = service.Spec.IPFamilies

	if service.Spec.Type == corev1.ServiceTypeLoadBalancer {
		if len(service.Status.LoadBalancer.Ingress) > 0 {
			for _, ingr := range service.Status.LoadBalancer.Ingress {
				s.externalIPs = append(s.externalIPs, ingr.IP)
			}
		} else {
			return ctrl.Result{}, fmt.Errorf("%s LoadBalancer IP still pending", s.service.Name)
		}
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
		err = fmt.Errorf("Error deleting service %s: %w", s.service.Name, err)
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
		err = fmt.Errorf("Error listing services for %s: %w", obj.GetName(), err)
		return err
	}

	// delete all pods
	for _, pod := range serviceList.Items {
		err := h.GetClient().Delete(ctx, &pod)
		if err != nil && !k8s_errors.IsNotFound(err) {
			err = fmt.Errorf("Error deleting service %s: %w", pod.Name, err)
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
		err = fmt.Errorf("Error listing services for labels: %v - %w", labelSelectorMap, err)
		return nil, err
	}

	return serviceList, nil
}

// GetServiceWithName - Get service with name in namespace
func GetServiceWithName(
	ctx context.Context,
	h *helper.Helper,
	name string,
	namespace string,
) (*corev1.Service, error) {

	service := &corev1.Service{}

	err := h.GetClient().Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, service)
	if err != nil {
		err = fmt.Errorf("Error getting service %s/%s - %w", name, namespace, err)

		return nil, err
	}

	return service, nil
}

// GetServicesPortDetails - Return ServicePort with name from a Service
func GetServicesPortDetails(
	service *corev1.Service,
	portName string,
) *corev1.ServicePort {

	for _, servicePort := range service.Spec.Ports {
		if servicePort.Name == portName {
			return &servicePort
		}
	}

	return nil
}

// EndptProtocol returns the protocol for the endpoint if proto is nil http is considered
func EndptProtocol(proto *Protocol) string {
	if proto == nil {
		return string(ProtocolHTTP) + "://"
	}
	if *proto == ProtocolNone {
		return ""
	}

	return string(*proto) + "://"
}
