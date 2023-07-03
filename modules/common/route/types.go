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

package route

import (
	"time"

	routev1 "github.com/openshift/api/route/v1"
)

// Route -
type Route struct {
	route    *routev1.Route
	timeout  time.Duration
	hostname string
	override *OverrideSpec
}

// GenericRouteDetails -
type GenericRouteDetails struct {
	Name           string
	Namespace      string
	Labels         map[string]string
	ServiceName    string
	TargetPortName string
	FQDN           string
}

// OverrideSpec configuration for the Route created to serve traffic to the cluster.
type OverrideSpec struct {
	// +optional
	*EmbeddedLabelsAnnotations `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	// Spec defines the behavior of a Route.
	// https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	//
	// The spec will be merged using StrategicMergePatch
	//   - Provided parameters will override the ones from the original spec.
	//   - Required parameters of sub structs have to be named.
	//   - For parameters which are list of struct it depends on the patchStrategy defined on the list
	//     https://kubernetes.io/docs/tasks/manage-kubernetes-objects/update-api-object-kubectl-patch/#notes-on-the-strategic-merge-patch
	//     If `patchStrategy:"merge"` is set, src and dst list gets merged, otherwise they get replaced.
	// +optional
	Spec *Spec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}

// EmbeddedLabelsAnnotations is an embedded subset of the fields included in k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta.
// Only labels and annotations are included.
// New labels/annotations get merged with the ones created by the operator. If a privided
// annotation/label is the same as one created by the service operator, the ones provided
// via this override will replace the one from the operator.
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

// Spec describes the hostname or path the route exposes, any security information,
// and one to four backends (services) the route points to. Requests are distributed
// among the backends depending on the weights assigned to each backend. When using
// roundrobin scheduling the portion of requests that go to each backend is the backend
// weight divided by the sum of all of the backend weights. When the backend has more than
// one endpoint the requests that end up on the backend are roundrobin distributed among
// the endpoints. Weights are between 0 and 256 with default 100. Weight 0 causes no requests
// to the backend. If all weights are zero the route will be considered to have no backends
// and return a standard 503 response.
//
// The `tls` field is optional and allows specific certificates or behavior for the
// route. Routers typically configure a default certificate on a wildcard domain to
// terminate routes without explicit certificates, but custom hostnames usually must
// choose passthrough (send traffic directly to the backend via the TLS Server-Name-
// Indication field) or provide a certificate.
//
// Copy of RouteSpec in https://github.com/openshift/api/blob/master/route/v1/types.go,
// parameters set to be optional, have omitempty, and no default.
type Spec struct {
	// host is an alias/DNS that points to the service. Optional.
	// If not specified a route name will typically be automatically
	// chosen.
	// Must follow DNS952 subdomain conventions.
	//
	// +optional
	// +kubebuilder:validation:MaxLength=253
	// +kubebuilder:validation:Pattern=`^([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])(\.([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]{0,61}[a-zA-Z0-9]))*$`
	Host string `json:"host,omitempty" protobuf:"bytes,1,opt,name=host"`
	// subdomain is a DNS subdomain that is requested within the ingress controller's
	// domain (as a subdomain). If host is set this field is ignored. An ingress
	// controller may choose to ignore this suggested name, in which case the controller
	// will report the assigned name in the status.ingress array or refuse to admit the
	// route. If this value is set and the server does not support this field host will
	// be populated automatically. Otherwise host is left empty. The field may have
	// multiple parts separated by a dot, but not all ingress controllers may honor
	// the request. This field may not be changed after creation except by a user with
	// the update routes/custom-host permission.
	//
	// Example: subdomain `frontend` automatically receives the router subdomain
	// `apps.mycluster.com` to have a full hostname `frontend.apps.mycluster.com`.
	//
	// +optional
	// +kubebuilder:validation:MaxLength=253
	// +kubebuilder:validation:Pattern=`^([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])(\.([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]{0,61}[a-zA-Z0-9]))*$`
	Subdomain string `json:"subdomain,omitempty" protobuf:"bytes,8,opt,name=subdomain"`

	// path that the router watches for, to route traffic for to the service. Optional
	//
	// +optional
	// +kubebuilder:validation:Pattern=`^/`
	Path string `json:"path,omitempty" protobuf:"bytes,2,opt,name=path"`

	// to is an object the route should use as the primary backend. Only the Service kind
	// is allowed, and it will be defaulted to Service. If the weight field (0-256 default 100)
	// is set to zero, no traffic will be sent to this backend.
	To TargetReference `json:"to,omitempty" protobuf:"bytes,3,opt,name=to"`

	// alternateBackends allows up to 3 additional backends to be assigned to the route.
	// Only the Service kind is allowed, and it will be defaulted to Service.
	// Use the weight field in RouteTargetReference object to specify relative preference.
	//
	// +kubebuilder:validation:MaxItems=3
	AlternateBackends []TargetReference `json:"alternateBackends,omitempty" protobuf:"bytes,4,rep,name=alternateBackends"`

	// If specified, the port to be used by the router. Most routers will use all
	// endpoints exposed by the service by default - set this value to instruct routers
	// which port to use.
	// +optional
	Port *routev1.RoutePort `json:"port,omitempty" protobuf:"bytes,5,opt,name=port"`

	// The tls field provides the ability to configure certificates and termination for the route.
	TLS *routev1.TLSConfig `json:"tls,omitempty" protobuf:"bytes,6,opt,name=tls"`

	// Wildcard policy if any for the route.
	// Currently only 'Subdomain' or 'None' is allowed.
	//
	// +kubebuilder:validation:Enum=None;Subdomain;""
	WildcardPolicy routev1.WildcardPolicyType `json:"wildcardPolicy,omitempty" protobuf:"bytes,7,opt,name=wildcardPolicy"`
}

// TargetReference specifies the target that resolve into endpoints. Only the 'Service'
// kind is allowed. Use 'weight' field to emphasize one over others.
// Copy of RouteTargetReference in https://github.com/openshift/api/blob/master/route/v1/types.go,
// parameters set to be optional, have omitempty, and no default.
type TargetReference struct {
	// The kind of target that the route is referring to. Currently, only 'Service' is allowed
	//
	// +optional
	// +kubebuilder:validation:Enum=Service;""
	Kind string `json:"kind,omitempty" protobuf:"bytes,1,opt,name=kind"`

	// name of the service/target that is being referred to. e.g. name of the service
	//
	// +optional
	Name string `json:"name,omitempty" protobuf:"bytes,2,opt,name=name"`

	// weight as an integer between 0 and 256, default 100, that specifies the target's relative weight
	// against other target reference objects. 0 suppresses requests to this backend.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=256
	Weight *int32 `json:"weight,omitempty" protobuf:"varint,3,opt,name=weight"`
}
