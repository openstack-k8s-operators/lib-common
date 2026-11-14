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

// Package networkpolicy provides utilities for managing Kubernetes NetworkPolicy resources
package networkpolicy

import (
	"context"
	"fmt"

	"github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	"github.com/openstack-k8s-operators/lib-common/modules/common/util"

	networkingv1 "k8s.io/api/networking/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// normalizedPolicyTypes returns n.PolicyTypes if explicitly set, otherwise
// computes the same default the Kubernetes API server itself applies when
// PolicyTypes is omitted: Ingress is always included, and Egress is added
// only when Egress rules are present. Normalizing here keeps the value we
// send explicit and stable, so CreateOrPatch never fights the API server's
// own defaulting (which would otherwise turn an omitted PolicyTypes into a
// persisted non-nil value, that this mutator would then overwrite back to
// nil on every subsequent reconcile, producing a perpetual no-op Patch).
func (n *NetworkPolicy) normalizedPolicyTypes() []networkingv1.PolicyType {
	if len(n.PolicyTypes) > 0 {
		return n.PolicyTypes
	}
	policyTypes := []networkingv1.PolicyType{networkingv1.PolicyTypeIngress}
	if len(n.Egress) > 0 {
		policyTypes = append(policyTypes, networkingv1.PolicyTypeEgress)
	}
	return policyTypes
}

// CreateOrPatch - creates or patches a NetworkPolicy, reconciles after Xs if
// object won't exist.
func (n *NetworkPolicy) CreateOrPatch(
	ctx context.Context,
	h *helper.Helper,
) (ctrl.Result, error) {
	networkPolicy := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      n.Name,
			Namespace: n.Namespace,
		},
	}

	op, err := controllerutil.CreateOrPatch(ctx, h.GetClient(), networkPolicy, func() error {
		networkPolicy.Labels = util.MergeStringMaps(n.Labels, networkPolicy.Labels)
		networkPolicy.Annotations = util.MergeStringMaps(n.Annotations, networkPolicy.Annotations)
		networkPolicy.Spec = networkingv1.NetworkPolicySpec{
			PodSelector: n.PodSelector,
			PolicyTypes: n.normalizedPolicyTypes(),
			Ingress:     n.Ingress,
			Egress:      n.Egress,
		}

		return controllerutil.SetControllerReference(h.GetBeforeObject(), networkPolicy, h.GetScheme())
	})
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			h.GetLogger().Info(fmt.Sprintf("NetworkPolicy %s not found, reconcile in %s", n.Name, n.Timeout))
			return ctrl.Result{RequeueAfter: n.Timeout}, nil
		}
		return ctrl.Result{}, err
	}
	if op != controllerutil.OperationResultNone {
		h.GetLogger().Info(fmt.Sprintf("NetworkPolicy %s - %s", n.Name, op))
	}

	return ctrl.Result{}, nil
}

// Delete - delete a NetworkPolicy.
func (n *NetworkPolicy) Delete(
	ctx context.Context,
	h *helper.Helper,
) error {
	networkPolicy := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      n.Name,
			Namespace: n.Namespace,
		},
	}
	err := h.GetClient().Delete(ctx, networkPolicy)
	if err != nil && !k8s_errors.IsNotFound(err) {
		return fmt.Errorf("error deleting NetworkPolicy %s: %w", n.Name, err)
	}
	return nil
}
