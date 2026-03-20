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

// AddConsumerFinalizer adds consumerFinalizer to the given object.
func AddConsumerFinalizer(
	ctx context.Context,
	h *helper.Helper,
	obj client.Object,
	consumerFinalizer string,
) error {
	before := obj.DeepCopyObject().(client.Object)
	if controllerutil.AddFinalizer(obj, consumerFinalizer) {
		patch := client.MergeFromWithOptions(before, client.MergeFromWithOptimisticLock{})
		if err := h.GetClient().Patch(ctx, obj, patch); err != nil {
			return fmt.Errorf("failed to add consumer finalizer to %s: %w", obj.GetName(), err)
		}
		h.GetLogger().Info("Added consumer finalizer", "object", obj.GetName(), "finalizer", consumerFinalizer)
	}
	return nil
}

// RemoveConsumerFinalizer removes consumerFinalizer from the given object.
func RemoveConsumerFinalizer(
	ctx context.Context,
	h *helper.Helper,
	obj client.Object,
	consumerFinalizer string,
) error {
	before := obj.DeepCopyObject().(client.Object)
	if controllerutil.RemoveFinalizer(obj, consumerFinalizer) {
		patch := client.MergeFromWithOptions(before, client.MergeFromWithOptimisticLock{})
		if err := h.GetClient().Patch(ctx, obj, patch); err != nil {
			return fmt.Errorf("failed to remove consumer finalizer from %s: %w", obj.GetName(), err)
		}
		h.GetLogger().Info("Removed consumer finalizer", "object", obj.GetName(), "finalizer", consumerFinalizer)
	}
	return nil
}

// ManageConsumerFinalizer adds consumerFinalizer to newObj and removes it from oldObj.
//
//	If both refer to the same object, returns early without mutating.
func ManageConsumerFinalizer(
	ctx context.Context,
	h *helper.Helper,
	newObj client.Object,
	oldObj client.Object,
	consumerFinalizer string,
) error {
	if newObj != nil && oldObj != nil &&
		newObj.GetNamespace() == oldObj.GetNamespace() &&
		newObj.GetName() == oldObj.GetName() {
		return nil
	}

	if newObj != nil {
		if err := AddConsumerFinalizer(ctx, h, newObj, consumerFinalizer); err != nil {
			return err
		}
	}

	if oldObj != nil {
		if err := RemoveConsumerFinalizer(ctx, h, oldObj, consumerFinalizer); err != nil {
			return err
		}
	}

	return nil
}
