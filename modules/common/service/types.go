/*
Copyright 2021 Red Hat

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

// +kubebuilder:object:generate:=true

package service

import (
	"time"

	corev1 "k8s.io/api/core/v1"
)

// Endpoint - typedef to enumerate Endpoint verbs
// NOTE: (mschuppert) have to duplicate this for now. Can not have
// circular dep back to endpoint pkg also can not move it at this
// point as we have circular dep in test module to keystone.
type Endpoint string

// Protocol of the endpoint (http/https)
type Protocol string

const (
	// EndpointAdmin - admin endpoint
	EndpointAdmin Endpoint = "admin"
	// EndpointInternal - internal endpoint
	EndpointInternal Endpoint = "internal"
	// EndpointPublic - public endpoint
	EndpointPublic Endpoint = "public"
	// AnnotationIngressCreateKey -
	AnnotationIngressCreateKey = "core.openstack.org/ingress_create"
	// AnnotationEndpointKey -
	AnnotationEndpointKey = "endpoint"
	// AnnotationHostnameKey -
	AnnotationHostnameKey = "dnsmasq.network.openstack.org/hostname"
	// ProtocolHTTP -
	ProtocolHTTP Protocol = "http"
	// ProtocolHTTPS -
	ProtocolHTTPS Protocol = "https"
	// ProtocolNone -
	ProtocolNone Protocol = ""
)

func (e *Endpoint) String() string {
	return string(*e)
}

func (p *Protocol) String() string {
	return string(*p)
}

// Service -
// +kubebuilder:object:generate:=false
type Service struct {
	service         *corev1.Service
	timeout         time.Duration
	clusterIPs      []string
	externalIPs     []string
	ipFamilies      []corev1.IPFamily
	serviceHostname string
}

// GenericServiceDetails -
// +kubebuilder:object:generate:=false
type GenericServiceDetails struct {
	Name      string
	Namespace string
	Labels    map[string]string
	Selector  map[string]string
	// deprecated, use Ports
	Port                     GenericServicePort
	Ports                    []corev1.ServicePort
	ClusterIP                string
	PublishNotReadyAddresses bool
}

// GenericServicePort -
// +kubebuilder:object:generate:=false
// NOTE: (mschuppert) deprecated, can be removed when service operators moved to Ports
type GenericServicePort struct {
	Name     string
	Port     int32
	Protocol corev1.Protocol // corev1.ProtocolTCP/ corev1.ProtocolUDP/ corev1.ProtocolSCTP - https://pkg.go.dev/k8s.io/api@v0.23.6/core/v1#Protocol
}

// MetalLBServiceDetails -
// +kubebuilder:object:generate:=false
type MetalLBServiceDetails struct {
	Name        string
	Namespace   string
	Annotations map[string]string
	Labels      map[string]string
	Selector    map[string]string
	// deprecated, use Ports
	Port  GenericServicePort
	Ports []corev1.ServicePort
}

const (
	// MetalLBAddressPoolAnnotation -
	MetalLBAddressPoolAnnotation = "metallb.universe.tf/address-pool"
	// MetalLBAllowSharedIPAnnotation -
	MetalLBAllowSharedIPAnnotation = "metallb.universe.tf/allow-shared-ip"
	// MetalLBLoadBalancerIPs -
	MetalLBLoadBalancerIPs = "metallb.universe.tf/loadBalancerIPs"
)

// OverrideSpec - service override configuration for the Service created to serve traffic to the cluster.
// Allows for the manifest of the created Service to be overwritten with custom configuration.
type OverrideSpec struct {
	*EmbeddedLabelsAnnotations `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Spec                       *OverrideServiceSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}

// RoutedOverrideSpec - a routed service override configuration for the Service created to serve traffic
// to the cluster. Allows for the manifest of the created Service to be overwritten with custom configuration.
type RoutedOverrideSpec struct {
	OverrideSpec `json:",inline"`
	EndpointURL  *string `json:"endpointURL,omitempty"`
}

// EmbeddedLabelsAnnotations is an embedded subset of the fields included in k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta.
// Only labels and annotations are included.
type EmbeddedLabelsAnnotations struct {
	// Map of string keys and values that can be used to organize and categorize
	// (scope and select) objects. May match selectors of replication controllers
	// and services.
	// More info: http://kubernetes.io/docs/user-guide/labels
	// +optional
	Labels map[string]string `json:"labels,omitempty" protobuf:"bytes,11,rep,name=labels"`

	// Annotations is an unstructured key value map stored with a resource that may be
	// set by external tools to store and retrieve arbitrary metadata. They are not
	// queryable and should be preserved when modifying objects.
	// More info: http://kubernetes.io/docs/user-guide/annotations
	// +optional
	Annotations map[string]string `json:"annotations,omitempty" protobuf:"bytes,12,rep,name=annotations"`
}

// OverrideServiceSpec is a subset of the fields included in https://pkg.go.dev/k8s.io/api@v0.26.6/core/v1#ServiceSpec
// Limited to Type, SessionAffinity, LoadBalancerSourceRanges, ExternalName, ExternalTrafficPolicy, SessionAffinityConfig,
// IPFamilyPolicy, LoadBalancerClass and InternalTrafficPolicy
type OverrideServiceSpec struct {
	// type determines how the Service is exposed. Defaults to ClusterIP. Valid
	// options are ExternalName, ClusterIP, NodePort, and LoadBalancer.
	// "ClusterIP" allocates a cluster-internal IP address for load-balancing
	// to endpoints. Endpoints are determined by the selector or if that is not
	// specified, by manual construction of an Endpoints object or
	// EndpointSlice objects. If clusterIP is "None", no virtual IP is
	// allocated and the endpoints are published as a set of endpoints rather
	// than a virtual IP.
	// "NodePort" builds on ClusterIP and allocates a port on every node which
	// routes to the same endpoints as the clusterIP.
	// "LoadBalancer" builds on NodePort and creates an external load-balancer
	// (if supported in the current cloud) which routes to the same endpoints
	// as the clusterIP.
	// "ExternalName" aliases this service to the specified externalName.
	// Several other fields do not apply to ExternalName services.
	// More info: https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types
	// +optional
	Type corev1.ServiceType `json:"type,omitempty" protobuf:"bytes,4,opt,name=type,casttype=ServiceType"`

	// Supports "ClientIP" and "None". Used to maintain session affinity.
	// Enable client IP based session affinity.
	// Must be ClientIP or None.
	// Defaults to None.
	// More info: https://kubernetes.io/docs/concepts/services-networking/service/#virtual-ips-and-service-proxies
	// +optional
	SessionAffinity corev1.ServiceAffinity `json:"sessionAffinity,omitempty" protobuf:"bytes,7,opt,name=sessionAffinity,casttype=ServiceAffinity"`

	// If specified and supported by the platform, this will restrict traffic through the cloud-provider
	// load-balancer will be restricted to the specified client IPs. This field will be ignored if the
	// cloud-provider does not support the feature."
	// More info: https://kubernetes.io/docs/tasks/access-application-cluster/create-external-load-balancer/
	// +optional
	LoadBalancerSourceRanges []string `json:"loadBalancerSourceRanges,omitempty" protobuf:"bytes,9,opt,name=loadBalancerSourceRanges"`

	// externalName is the external reference that discovery mechanisms will
	// return as an alias for this service (e.g. a DNS CNAME record). No
	// proxying will be involved.  Must be a lowercase RFC-1123 hostname
	// (https://tools.ietf.org/html/rfc1123) and requires `type` to be "ExternalName".
	// +optional
	ExternalName string `json:"externalName,omitempty" protobuf:"bytes,10,opt,name=externalName"`

	// externalTrafficPolicy describes how nodes distribute service traffic they
	// receive on one of the Service's "externally-facing" addresses (NodePorts,
	// ExternalIPs, and LoadBalancer IPs). If set to "Local", the proxy will configure
	// the service in a way that assumes that external load balancers will take care
	// of balancing the service traffic between nodes, and so each node will deliver
	// traffic only to the node-local endpoints of the service, without masquerading
	// the client source IP. (Traffic mistakenly sent to a node with no endpoints will
	// be dropped.) The default value, "Cluster", uses the standard behavior of
	// routing to all endpoints evenly (possibly modified by topology and other
	// features). Note that traffic sent to an External IP or LoadBalancer IP from
	// within the cluster will always get "Cluster" semantics, but clients sending to
	// a NodePort from within the cluster may need to take traffic policy into account
	// when picking a node.
	// +optional
	ExternalTrafficPolicy corev1.ServiceExternalTrafficPolicyType `json:"externalTrafficPolicy,omitempty" protobuf:"bytes,11,opt,name=externalTrafficPolicy"`

	// sessionAffinityConfig contains the configurations of session affinity.
	// +optional
	SessionAffinityConfig *corev1.SessionAffinityConfig `json:"sessionAffinityConfig,omitempty" protobuf:"bytes,14,opt,name=sessionAffinityConfig"`

	// IPFamilyPolicy represents the dual-stack-ness requested or required by
	// this Service. If there is no value provided, then this field will be set
	// to SingleStack. Services can be "SingleStack" (a single IP family),
	// "PreferDualStack" (two IP families on dual-stack configured clusters or
	// a single IP family on single-stack clusters), or "RequireDualStack"
	// (two IP families on dual-stack configured clusters, otherwise fail). The
	// ipFamilies and clusterIPs fields depend on the value of this field. This
	// field will be wiped when updating a service to type ExternalName.
	// +optional
	IPFamilyPolicy *corev1.IPFamilyPolicy `json:"ipFamilyPolicy,omitempty" protobuf:"bytes,17,opt,name=ipFamilyPolicy,casttype=IPFamilyPolicy"`

	// loadBalancerClass is the class of the load balancer implementation this Service belongs to.
	// If specified, the value of this field must be a label-style identifier, with an optional prefix,
	// e.g. "internal-vip" or "example.com/internal-vip". Unprefixed names are reserved for end-users.
	// This field can only be set when the Service type is 'LoadBalancer'. If not set, the default load
	// balancer implementation is used, today this is typically done through the cloud provider integration,
	// but should apply for any default implementation. If set, it is assumed that a load balancer
	// implementation is watching for Services with a matching class. Any default load balancer
	// implementation (e.g. cloud providers) should ignore Services that set this field.
	// This field can only be set when creating or updating a Service to type 'LoadBalancer'.
	// Once set, it can not be changed. This field will be wiped when a service is updated to a non 'LoadBalancer' type.
	// +optional
	LoadBalancerClass *string `json:"loadBalancerClass,omitempty" protobuf:"bytes,21,opt,name=loadBalancerClass"`

	// InternalTrafficPolicy describes how nodes distribute service traffic they
	// receive on the ClusterIP. If set to "Local", the proxy will assume that pods
	// only want to talk to endpoints of the service on the same node as the pod,
	// dropping the traffic if there are no local endpoints. The default value,
	// "Cluster", uses the standard behavior of routing to all endpoints evenly
	// (possibly modified by topology and other features).
	// +optional
	InternalTrafficPolicy *corev1.ServiceInternalTrafficPolicyType `json:"internalTrafficPolicy,omitempty" protobuf:"bytes,22,opt,name=internalTrafficPolicy"`
}
