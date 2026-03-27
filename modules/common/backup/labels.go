/*
Copyright 2025 Red Hat

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

// Package backup provides utilities for backup and restore labeling
package backup

import (
	"context"
	"fmt"
	"strings"

	k8s_corev1 "k8s.io/api/core/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	// BackupRestoreLabel is a CRD label: "true" means instances participate in backup/restore
	BackupRestoreLabel = "backup.openstack.org/restore"
	// BackupCategoryLabel is a CRD & instance label: "controlplane" or "dataplane"
	BackupCategoryLabel = "backup.openstack.org/category"
	// BackupRestoreOrderLabel is a CRD & instance label: "00"-"60"
	BackupRestoreOrderLabel = "backup.openstack.org/restore-order"

	// BackupLabel is a resource instance label: "true" marks for backup
	BackupLabel = "backup.openstack.org/backup"
)

// LabelKeys returns all backup-related label/annotation keys.
// Used by applyAnnotationOverrides and can be used by controllers
// for event filtering (e.g. detecting annotation changes).
func LabelKeys() []string {
	return []string{BackupLabel, BackupRestoreLabel, BackupRestoreOrderLabel}
}

// GetBackupLabels returns labels to mark a resource for OADP backup selection.
// Use this for PVCs and other resources that need to be explicitly selected
// for backup (large storage volumes). Resources backed up by namespace
// (CRs, Secrets, ConfigMaps) do not need these labels.
func GetBackupLabels(category string) map[string]string {
	labels := map[string]string{
		BackupLabel: "true",
	}
	if category != "" {
		labels[BackupCategoryLabel] = category
	}
	return labels
}

// GetRestoreLabels returns labels for controlling restore ordering.
// Use this for CRs, Secrets, ConfigMaps that are backed up by namespace
// but need ordered restore. For PVCs, combine with GetBackupLabels().
func GetRestoreLabels(restoreOrder, category string) map[string]string {
	labels := map[string]string{
		BackupRestoreLabel:      "true",
		BackupRestoreOrderLabel: restoreOrder,
	}
	if category != "" {
		labels[BackupCategoryLabel] = category
	}
	return labels
}

// GetRestoreLabelsWithOverrides returns restore labels with overrides.
// The overrides map (typically from CR annotations) can override the default
// restoreOrder and category.
func GetRestoreLabelsWithOverrides(defaultRestoreOrder string, overrides map[string]string) map[string]string {
	labels := GetRestoreLabels(defaultRestoreOrder, "")

	// Check for user override of restore order
	if order, ok := overrides[BackupRestoreOrderLabel]; ok {
		labels[BackupRestoreOrderLabel] = order
	}

	// Category override
	if category, ok := overrides[BackupCategoryLabel]; ok {
		labels[BackupCategoryLabel] = category
	}

	return labels
}

// ShouldBackup returns true if the resource is marked for backup
func ShouldBackup(labels map[string]string) bool {
	return labels != nil && labels[BackupLabel] == "true"
}

// EnsureBackupLabels sets backup/restore labels on a resource. It always
// writes the caller-provided default labels, then applies any annotation
// overrides on top. This means:
//   - Operator defaults are always current (updated on operator upgrade)
//   - User overrides via annotations take precedence
//   - It's visible what was set by the operator vs what the user overrode
//
// defaultLabels should be built by the caller using GetBackupLabels() and/or
// GetRestoreLabels(). Returns true if labels were changed.
func EnsureBackupLabels(ctx context.Context, c client.Client, obj client.Object, defaultLabels map[string]string) (bool, error) {
	origObj := obj.DeepCopyObject().(client.Object)

	labels := obj.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}

	// Step 1: Always set operator defaults (overwrite existing)
	for k, v := range defaultLabels {
		labels[k] = v
	}

	// Step 2: Apply annotation overrides on top
	ApplyAnnotationOverrides(obj.GetAnnotations(), labels)

	// Check if anything actually changed
	origLabels := origObj.GetLabels()
	changed := len(labels) != len(origLabels)
	if !changed {
		for k, v := range labels {
			if origLabels[k] != v {
				changed = true
				break
			}
		}
	}
	if !changed {
		return false, nil
	}

	patch := client.MergeFrom(origObj)
	obj.SetLabels(labels)
	if err := c.Patch(ctx, obj, patch); err != nil {
		return false, fmt.Errorf("patching backup labels on %s: %w", obj.GetName(), err)
	}
	return true, nil
}

// GetCertSecretBackupLabels returns backup labels for a cert secret, respecting
// annotation overrides. It reads the cert secret (named "cert-<certName>") and
// checks for backup-related annotations. If found, they override the default labels.
// This ensures that cert-manager's SecretTemplate propagates the correct labels
// so that annotation overrides on the Secret are not reverted by cert-manager.
func GetCertSecretBackupLabels(
	ctx context.Context,
	c client.Client,
	certName string,
	namespace string,
	defaultLabels map[string]string,
) (map[string]string, error) {
	labels := make(map[string]string, len(defaultLabels))
	for k, v := range defaultLabels {
		labels[k] = v
	}

	// Check if the cert secret already exists and has annotation overrides
	certSecretName := "cert-" + certName
	certSecret := &k8s_corev1.Secret{}
	if err := c.Get(ctx, types.NamespacedName{Name: certSecretName, Namespace: namespace}, certSecret); err != nil {
		if !k8s_errors.IsNotFound(err) {
			return nil, fmt.Errorf("failed to get cert secret %s/%s: %w", namespace, certSecretName, err)
		}
		// Secret doesn't exist yet (first reconcile) — use defaults
		return labels, nil
	}

	// Apply annotation overrides from the secret
	ApplyAnnotationOverrides(certSecret.GetAnnotations(), labels)
	return labels, nil
}

// AnnotationChangedPredicate returns a predicate that only triggers
// for resources matching the given label selector when backup annotations change.
// This is useful for watching resources (e.g. cert secrets) where a user may
// add backup annotation overrides that need to be picked up by a controller.
func AnnotationChangedPredicate(labelSelector string) predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return false
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			labels := e.ObjectNew.GetLabels()
			if _, ok := labels[labelSelector]; !ok {
				return false
			}

			oldAnnotations := e.ObjectOld.GetAnnotations()
			newAnnotations := e.ObjectNew.GetAnnotations()
			for _, key := range LabelKeys() {
				if oldAnnotations[key] != newAnnotations[key] {
					return true
				}
			}
			return false
		},
	}
}

// ApplyAnnotationOverrides checks for backup-related annotations on a resource
// and applies them as label overrides. Annotations allow users to override
// operator defaults:
//   - backup.openstack.org/backup: "false" → exclude from backup
//   - backup.openstack.org/restore: "false" → skip restore
//   - backup.openstack.org/restore-order: "XX" → custom restore order (implies restore=true)
func ApplyAnnotationOverrides(annotations, labels map[string]string) {
	if annotations == nil {
		return
	}

	for _, key := range LabelKeys() {
		val, has := annotations[key]
		if !has {
			continue
		}
		normalized := strings.ToLower(val)

		switch key {
		case BackupLabel:
			labels[BackupLabel] = normalized
		case BackupRestoreLabel:
			labels[BackupRestoreLabel] = normalized
		case BackupRestoreOrderLabel:
			labels[BackupRestoreOrderLabel] = normalized
			labels[BackupRestoreLabel] = "true"
		}
	}
}
