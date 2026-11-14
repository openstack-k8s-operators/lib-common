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
	"context"
	"testing"
	"time"

	. "github.com/onsi/gomega" // nolint:revive
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/openstack-k8s-operators/lib-common/modules/common/helper"
)

var timeout = 5 * time.Second

func TestNormalizedPolicyTypes(t *testing.T) {
	tests := []struct {
		name string
		np   NetworkPolicy
		want []networkingv1.PolicyType
	}{
		{
			name: "explicit PolicyTypes is returned unchanged",
			np: NetworkPolicy{
				PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeEgress},
				Ingress:     []networkingv1.NetworkPolicyIngressRule{{}},
			},
			want: []networkingv1.PolicyType{networkingv1.PolicyTypeEgress},
		},
		{
			name: "omitted PolicyTypes with no rules defaults to Ingress",
			np:   NetworkPolicy{},
			want: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress},
		},
		{
			name: "omitted PolicyTypes with only Ingress rules defaults to Ingress",
			np: NetworkPolicy{
				Ingress: []networkingv1.NetworkPolicyIngressRule{{}},
			},
			want: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress},
		},
		{
			name: "omitted PolicyTypes with Egress rules defaults to Ingress and Egress",
			np: NetworkPolicy{
				Egress: []networkingv1.NetworkPolicyEgressRule{{}},
			},
			want: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress, networkingv1.PolicyTypeEgress},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(tt.np.normalizedPolicyTypes()).To(Equal(tt.want))
		})
	}
}

func TestNetworkPolicyCreateOrPatch(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	owner := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "owner", Namespace: "namespace", UID: "owner-uid"}}

	client := fake.NewClientBuilder().WithScheme(scheme.Scheme).Build()
	h, err := helper.NewHelper(owner, client, nil, scheme.Scheme, ctrl.Log)
	g.Expect(err).NotTo(HaveOccurred())

	np := &NetworkPolicy{
		Name:      "foo",
		Namespace: "namespace",
		Labels:    map[string]string{"foo": "bar"},
		PodSelector: metav1.LabelSelector{
			MatchLabels: map[string]string{"service": "servicename"},
		},
		PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress},
		Ingress: []networkingv1.NetworkPolicyIngressRule{
			{
				From: []networkingv1.NetworkPolicyPeer{
					{PodSelector: &metav1.LabelSelector{}},
				},
			},
		},
		Timeout: timeout,
	}

	// create network policy
	_, err = np.CreateOrPatch(ctx, h)
	g.Expect(err).NotTo(HaveOccurred())

	created := &networkingv1.NetworkPolicy{}
	g.Expect(client.Get(ctx, types.NamespacedName{Name: "foo", Namespace: "namespace"}, created)).To(Succeed())
	g.Expect(created.Labels).To(Equal(map[string]string{"foo": "bar"}))
	g.Expect(created.Spec.PodSelector.MatchLabels).To(Equal(map[string]string{"service": "servicename"}))
	g.Expect(created.Spec.PolicyTypes).To(Equal([]networkingv1.PolicyType{networkingv1.PolicyTypeIngress}))
	g.Expect(created.Spec.Ingress).To(HaveLen(1))
	g.Expect(created.OwnerReferences).To(HaveLen(1))
	g.Expect(created.OwnerReferences[0].Name).To(Equal("owner"))

	// update: change the PodSelector and reconcile again
	np.PodSelector = metav1.LabelSelector{
		MatchLabels: map[string]string{"service": "servicename", "owner": "ownername"},
	}
	_, err = np.CreateOrPatch(ctx, h)
	g.Expect(err).NotTo(HaveOccurred())

	updated := &networkingv1.NetworkPolicy{}
	g.Expect(client.Get(ctx, types.NamespacedName{Name: "foo", Namespace: "namespace"}, updated)).To(Succeed())
	g.Expect(updated.Spec.PodSelector.MatchLabels).To(Equal(map[string]string{"service": "servicename", "owner": "ownername"}))
}

func TestNetworkPolicyDelete(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	owner := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "owner", Namespace: "namespace", UID: "owner-uid"}}

	client := fake.NewClientBuilder().WithScheme(scheme.Scheme).Build()
	h, err := helper.NewHelper(owner, client, nil, scheme.Scheme, ctrl.Log)
	g.Expect(err).NotTo(HaveOccurred())

	np := &NetworkPolicy{
		Name:      "foo",
		Namespace: "namespace",
		Timeout:   timeout,
	}
	_, err = np.CreateOrPatch(ctx, h)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(np.Delete(ctx, h)).To(Succeed())

	deleted := &networkingv1.NetworkPolicy{}
	err = client.Get(ctx, types.NamespacedName{Name: "foo", Namespace: "namespace"}, deleted)
	g.Expect(k8s_errors.IsNotFound(err)).To(BeTrue())

	// deleting again is idempotent
	g.Expect(np.Delete(ctx, h)).To(Succeed())
}
