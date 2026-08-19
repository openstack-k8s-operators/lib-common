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
	"time"

	commonannotations "github.com/openstack-k8s-operators/lib-common/modules/common/annotations"
	"github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	corev1 "k8s.io/api/core/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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

// ManageSecretConsumerFinalizer ensures consumerFinalizer is present on the
// secret identified by secretName. It is a no-op when secretName is empty.
func ManageSecretConsumerFinalizer(
	ctx context.Context,
	h *helper.Helper,
	namespace string,
	secretName string,
	consumerFinalizer string,
) error {
	if secretName == "" {
		return nil
	}

	secret := &corev1.Secret{}
	key := types.NamespacedName{Name: secretName, Namespace: namespace}
	if err := h.GetClient().Get(ctx, key, secret); err != nil {
		return fmt.Errorf("failed to get secret %s: %w", secretName, err)
	}

	return AddConsumerFinalizer(ctx, h, secret, consumerFinalizer)
}

// RemoveSecretConsumerFinalizer removes consumerFinalizer from the secret
// identified by secretName. It is a no-op when secretName is empty or the
// secret no longer exists.
func RemoveSecretConsumerFinalizer(
	ctx context.Context,
	h *helper.Helper,
	namespace string,
	secretName string,
	consumerFinalizer string,
) error {
	if secretName == "" {
		return nil
	}

	secret := &corev1.Secret{}
	key := types.NamespacedName{Name: secretName, Namespace: namespace}
	if err := h.GetClient().Get(ctx, key, secret); err != nil {
		if k8s_errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to get secret %s: %w", secretName, err)
	}
	return RemoveConsumerFinalizer(ctx, h, secret, consumerFinalizer)
}

// PruneSecretConsumerFinalizers removes consumerFinalizer from every secret in
// namespace that still carries it and is not named in keep. Empty names in keep
// are ignored.
//
// It self-heals finalizers stranded on superseded secrets: during rapid rotation
// (e.g. secret A -> B -> C before the workload becomes ready), the consumer
// finalizer can be added to an intermediate secret B that the workload never
// applies. FinalizeSecretRotation only ever releases the single tracked "old"
// secret, so B's finalizer would otherwise leak and block deletion of its Secret
// and revocation of the associated backend credential. Callers pass the set of
// secrets that legitimately hold the finalizer (the applied and current secrets
// for every rotating input guarded by the same finalizer) as keep, and any other
// secret carrying it is pruned.
//
// The secrets are read through the caller's (cached) client; the prune is
// idempotent and safe to call every reconcile.
//
// keep must enumerate every secret in the namespace that legitimately holds
// consumerFinalizer: the prune removes it from all others. This assumes a single
// consumer of consumerFinalizer per namespace (the CR is a namespace singleton,
// as is the norm for OpenStackControlPlane-managed services). If multiple CR
// instances sharing the same consumerFinalizer coexist in one namespace, each
// only knows its own keep set and would prune the others' finalizers; use a
// per-instance finalizer in that case.
func PruneSecretConsumerFinalizers(
	ctx context.Context,
	h *helper.Helper,
	namespace string,
	consumerFinalizer string,
	keep ...string,
) error {
	keepSet := make(map[string]struct{}, len(keep))
	for _, name := range keep {
		if name != "" {
			keepSet[name] = struct{}{}
		}
	}

	secretList := &corev1.SecretList{}
	if err := h.GetClient().List(ctx, secretList, client.InNamespace(namespace)); err != nil {
		return fmt.Errorf("failed to list secrets in namespace %s: %w", namespace, err)
	}

	for i := range secretList.Items {
		secret := &secretList.Items[i]
		if _, ok := keepSet[secret.Name]; ok {
			continue
		}
		if !controllerutil.ContainsFinalizer(secret, consumerFinalizer) {
			continue
		}
		if err := RemoveConsumerFinalizer(ctx, h, secret, consumerFinalizer); err != nil {
			return err
		}
	}

	return nil
}

// FinalizeSecretRotation handles the rotation guard for a credential secret
// (transport URL, application credential, or any other rotating secret).
// It detects whether rotation is in progress and:
//   - If no rotation (statusSecretName == currentSecretName or statusSecretName
//     is empty): returns currentSecretName
//   - If rotation detected and guardReady is true: removes the consumer
//     finalizer from the old secret and returns currentSecretName
//   - If rotation detected and guardReady is false: returns
//     statusSecretName unchanged (finalizer held, rotation pending)
//
// Compute guardReady with condition.CredentialRotationGuardReady from lib-common.
func FinalizeSecretRotation(
	ctx context.Context,
	h *helper.Helper,
	namespace string,
	statusSecretName string,
	currentSecretName string,
	consumerFinalizer string,
	guardReady bool,
) (string, error) {
	if statusSecretName == "" || statusSecretName == currentSecretName {
		return currentSecretName, nil
	}

	if !guardReady {
		return statusSecretName, nil
	}

	if err := RemoveSecretConsumerFinalizer(
		ctx, h, namespace, statusSecretName, consumerFinalizer,
	); err != nil {
		return statusSecretName, err
	}
	return currentSecretName, nil
}

// ManageRotationGracePeriod manages a time-based grace period during
// credential rotation. When rotationPending is true and no grace period
// is active, it sets a timestamp annotation and returns (requeue, true, nil).
// While the grace period is active, it returns (requeue, true, nil) with
// the remaining time. After the grace period expires, it returns
// ({}, false, nil) so the caller can evaluate the rotation guard.
// When rotationPending is false, it clears any existing annotation.
//
// This gives sub-CRs time to detect config changes, update their
// Deployments/StatefulSets, and roll pods before the guard releases
// the old secret's consumer finalizer.
func ManageRotationGracePeriod(
	ctx context.Context,
	c client.Client,
	obj client.Object,
	rotationPending bool,
	gracePeriod time.Duration,
) (ctrl.Result, bool, error) {
	annotations := obj.GetAnnotations()

	if !rotationPending {
		if annotations != nil {
			if _, has := annotations[commonannotations.RotationGraceAnnotation]; has {
				before := obj.DeepCopyObject().(client.Object)
				delete(annotations, commonannotations.RotationGraceAnnotation)
				obj.SetAnnotations(annotations)
				if err := c.Patch(ctx, obj, client.MergeFrom(before)); err != nil {
					return ctrl.Result{}, false, err
				}
			}
		}
		return ctrl.Result{}, false, nil
	}

	graceUntilStr := ""
	if annotations != nil {
		graceUntilStr = annotations[commonannotations.RotationGraceAnnotation]
	}

	if graceUntilStr == "" {
		before := obj.DeepCopyObject().(client.Object)
		graceUntil := time.Now().Add(gracePeriod)
		if annotations == nil {
			annotations = map[string]string{}
		}
		annotations[commonannotations.RotationGraceAnnotation] = graceUntil.Format(time.RFC3339)
		obj.SetAnnotations(annotations)
		if err := c.Patch(ctx, obj, client.MergeFrom(before)); err != nil {
			return ctrl.Result{}, false, err
		}
		return ctrl.Result{RequeueAfter: gracePeriod}, true, nil
	}

	graceUntil, err := time.Parse(time.RFC3339, graceUntilStr)
	if err != nil {
		before := obj.DeepCopyObject().(client.Object)
		delete(annotations, commonannotations.RotationGraceAnnotation)
		obj.SetAnnotations(annotations)
		if patchErr := c.Patch(ctx, obj, client.MergeFrom(before)); patchErr != nil {
			return ctrl.Result{}, false, patchErr
		}
		return ctrl.Result{}, false, nil
	}

	if time.Now().Before(graceUntil) {
		remaining := time.Until(graceUntil)
		return ctrl.Result{RequeueAfter: remaining}, true, nil
	}

	// Grace period expired — let the caller evaluate the guard.
	// The annotation stays until rotationPending becomes false
	// (after FinalizeSecretRotation updates the status), at which
	// point the !rotationPending branch above clears it.
	return ctrl.Result{}, false, nil
}
