/*
Copyright 2026 Red Hat

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

package networkpolicy

import (
	"time"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NetworkPolicy -
type NetworkPolicy struct {
	Name        string
	Namespace   string
	Labels      map[string]string
	Annotations map[string]string
	PodSelector metav1.LabelSelector
	PolicyTypes []networkingv1.PolicyType
	Ingress     []networkingv1.NetworkPolicyIngressRule
	Egress      []networkingv1.NetworkPolicyEgressRule
	Timeout     time.Duration
}
