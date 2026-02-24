/*
Copyright 2023 Red Hat

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

// Package object provides utilities for managing Kubernetes object metadata and operations
package object

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// CheckOwnerRefExist - returns true if the owner is already in the owner ref list
func CheckOwnerRefExist(
	uid types.UID,
	ownerRefs []metav1.OwnerReference,
) bool {
	f := func(o metav1.OwnerReference) bool {
		return o.UID == uid
	}
	if idx := slices.IndexFunc(ownerRefs, f); idx >= 0 {
		return true
	}

	return false
}

// PatchOwnerRef - creates a patch to add ownerref to an object
func PatchOwnerRef(
	owner client.Object,
	object client.Object,
	scheme *runtime.Scheme,
) (map[string]interface{}, client.Patch, error) {
	beforeObject := object.DeepCopyObject().(client.Object)

	// add owner ref to the object
	err := controllerutil.SetOwnerReference(owner, object, scheme)
	if err != nil {
		return nil, nil, err
	}

	// create patch
	patch := client.MergeFrom(beforeObject)
	diff, err := patch.Data(object)
	if err != nil {
		return nil, nil, err
	}

	// Unmarshal patch data into a local map for logging
	patchDiff := map[string]interface{}{}
	if err := json.Unmarshal(diff, &patchDiff); err != nil {
		return nil, nil, err
	}

	return patchDiff, patch, nil
}

// EnsureOwnerRef - adds owner ref (no controller) to an object which then can
// can be used to reconcile when the object changes by adding the following in
// NewControllerManagedBy().
// Note: This will not triggere a reconcilation when the object gets re-created
// from scratch, like deleting a secret.
//
// watch for secrets we added ourselves as additional owners, NOT as controller
// Watches(
//
//	&source.Kind{Type: &corev1.Secret{}},
//	&handler.EnqueueRequestForOwner{OwnerType: &clientv1.OpenStackClient{}, IsController: false}).
func EnsureOwnerRef(
	ctx context.Context,
	h *helper.Helper,
	owner client.Object,
	object client.Object,
) error {
	// create owner ref patch
	patchDiff, patch, err := PatchOwnerRef(owner, object, h.GetScheme())
	if err != nil {
		return err
	}

	if _, ok := patchDiff["metadata"]; ok {
		err = h.GetClient().Patch(ctx, object, patch)
		if k8s_errors.IsConflict(err) {
			return fmt.Errorf("error metadata update conflict: %w", err)
		} else if err != nil && !k8s_errors.IsNotFound(err) {
			return fmt.Errorf("error metadata update failed: %w", err)
		}

		h.GetLogger().Info(fmt.Sprintf("Owner reference patched - diff %+v", patchDiff["metadata"]))
	}

	return nil
}

// IsOwnerServiceReady checks if the owner service that owns this object is ready.
// Returns true if the owner is ready, false if not ready, and error only for unexpected failures.
// If there's no owner with controller=true, it returns true (safe to proceed).
func IsOwnerServiceReady(
	ctx context.Context,
	h *helper.Helper,
	obj client.Object,
) (bool, error) {
	// Find the controller owner reference (e.g., Cinder, Nova, etc.)
	var ownerRef *metav1.OwnerReference
	for _, owner := range obj.GetOwnerReferences() {
		if owner.Controller != nil && *owner.Controller {
			ownerRef = &owner
			break
		}
	}

	// If no controlling owner, safe to proceed
	if ownerRef == nil {
		h.GetLogger().Info("No controller owner found, owner is considered ready")
		return true, nil
	}

	// Parse the APIVersion to extract group and version
	gv, err := schema.ParseGroupVersion(ownerRef.APIVersion)
	if err != nil {
		h.GetLogger().Error(err, "Failed to parse owner APIVersion", "apiVersion", ownerRef.APIVersion)
		return false, err
	}

	// Fetch the owner resource using unstructured client
	owner := &unstructured.Unstructured{}
	owner.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   gv.Group,
		Version: gv.Version,
		Kind:    ownerRef.Kind,
	})

	err = h.GetClient().Get(ctx, types.NamespacedName{
		Name:      ownerRef.Name,
		Namespace: obj.GetNamespace(),
	}, owner)

	if err != nil {
		if k8s_errors.IsNotFound(err) {
			// Owner deleted, safe to proceed
			h.GetLogger().Info("Owner resource not found, owner is considered ready", "kind", ownerRef.Kind, "name", ownerRef.Name)
			return true, nil
		}
		// Unexpected error, log and return error
		h.GetLogger().Error(err, "Failed to fetch owner resource", "kind", ownerRef.Kind, "name", ownerRef.Name)
		return false, err
	}

	// Check status.conditions for Ready condition
	conditions, found, err := unstructured.NestedSlice(owner.Object, "status", "conditions")
	if err != nil || !found {
		h.GetLogger().Info("No conditions found in owner status, waiting", "kind", ownerRef.Kind, "name", ownerRef.Name)
		return false, nil
	}

	// Look for Ready condition with status=True
	isReady := false
	for _, c := range conditions {
		condition, ok := c.(map[string]any)
		if !ok {
			continue
		}

		condType, _, _ := unstructured.NestedString(condition, "type")
		status, _, _ := unstructured.NestedString(condition, "status")

		if condType == "Ready" && status == "True" {
			isReady = true
			break
		}
	}

	if !isReady {
		h.GetLogger().Info("Owner service not ready, waiting", "kind", ownerRef.Kind, "name", ownerRef.Name)
		return false, nil
	}

	// Check if owner has reconciled (observedGeneration matches generation)
	generation, foundGen, err := unstructured.NestedInt64(owner.Object, "metadata", "generation")
	if err != nil || !foundGen {
		h.GetLogger().Info("Could not get owner generation, waiting", "kind", ownerRef.Kind, "name", ownerRef.Name)
		return false, nil
	}

	observedGeneration, foundObsGen, err := unstructured.NestedInt64(owner.Object, "status", "observedGeneration")
	if err != nil || !foundObsGen {
		h.GetLogger().Info("Could not get owner observedGeneration, waiting", "kind", ownerRef.Kind, "name", ownerRef.Name)
		return false, nil
	}

	if observedGeneration != generation {
		h.GetLogger().Info("Owner service has not reconciled yet (observedGeneration != generation), waiting",
			"kind", ownerRef.Kind,
			"name", ownerRef.Name,
			"generation", generation,
			"observedGeneration", observedGeneration)
		return false, nil
	}

	h.GetLogger().Info("Owner service is ready and has reconciled, safe to proceed", "kind", ownerRef.Kind, "name", ownerRef.Name)
	return true, nil
}
