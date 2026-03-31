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

// Package statefulset provides utilities for managing Kubernetes StatefulSet resources
package statefulset

import (
	"context"
	"fmt"
	"time"

	"github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	"github.com/openstack-k8s-operators/lib-common/modules/common/util"
	appsv1 "k8s.io/api/apps/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// NewStatefulSet returns an initialized NewStatefulset.
func NewStatefulSet(
	statefulset *appsv1.StatefulSet,
	timeout time.Duration,
) *StatefulSet {
	return &StatefulSet{
		statefulset: statefulset,
		timeout:     timeout,
	}
}

// CreateOrPatch - creates or patches a statefulset, reconciles after Xs if object won't exist.
func (s *StatefulSet) CreateOrPatch(
	ctx context.Context,
	h *helper.Helper,
) (ctrl.Result, error) {
	statefulset := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.statefulset.Name,
			Namespace: s.statefulset.Namespace,
		},
	}

	op, err := controllerutil.CreateOrPatch(ctx, h.GetClient(), statefulset, func() error {
		statefulset.Labels = util.MergeStringMaps(statefulset.Labels, s.statefulset.Labels)
		statefulset.Annotations = util.MergeStringMaps(statefulset.Annotations, s.statefulset.Annotations)

		// Selector and VolumeClaimTemplates are immutable after creation.
		// Preserve the existing values so the full Spec overwrite below
		// does not trigger an API error on update.
		if !statefulset.CreationTimestamp.IsZero() {
			s.statefulset.Spec.Selector = statefulset.Spec.Selector
			s.statefulset.Spec.VolumeClaimTemplates = statefulset.Spec.VolumeClaimTemplates
		}

		// Save existing containers before overwriting the Spec so we can
		// merge them below to preserve server-defaulted fields.
		existingContainers := statefulset.Spec.Template.Spec.Containers
		existingInitContainers := statefulset.Spec.Template.Spec.InitContainers

		// Overwrite the entire Spec with the desired state. This ensures
		// any new Kubernetes fields are picked up automatically without
		// needing to add individual field copies.
		statefulset.Spec = s.statefulset.Spec

		// Merge containers by name to preserve server-defaulted fields
		// (e.g. TerminationMessagePath, ImagePullPolicy) and avoid
		// unnecessary reconcile loops. Falls back to full replacement if
		// container sets don't match by name.
		statefulset.Spec.Template.Spec.Containers = existingContainers
		MergeContainersByName(
			&statefulset.Spec.Template.Spec.Containers,
			s.statefulset.Spec.Template.Spec.Containers,
		)
		statefulset.Spec.Template.Spec.InitContainers = existingInitContainers
		MergeContainersByName(
			&statefulset.Spec.Template.Spec.InitContainers,
			s.statefulset.Spec.Template.Spec.InitContainers,
		)

		return controllerutil.SetControllerReference(h.GetBeforeObject(), statefulset, h.GetScheme())
	})
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			h.GetLogger().Info(fmt.Sprintf("StatefulSet %s not found, reconcile in %s", statefulset.Name, s.timeout))
			return ctrl.Result{RequeueAfter: s.timeout}, nil
		}
		return ctrl.Result{}, err
	}
	if op != controllerutil.OperationResultNone {
		h.GetLogger().Info(fmt.Sprintf("StatefulSet %s - %s", statefulset.Name, op))
	}

	// update the statefulset object of the statefulset type
	s.statefulset, err = GetStatefulSetWithName(ctx, h, statefulset.GetName(), statefulset.GetNamespace())
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// GetStatefulSet - get the statefulset object.
func (s *StatefulSet) GetStatefulSet() appsv1.StatefulSet {
	return *s.statefulset
}

// GetStatefulSetWithName func
func GetStatefulSetWithName(
	ctx context.Context,
	h *helper.Helper,
	name string,
	namespace string,
) (*appsv1.StatefulSet, error) {

	depl := &appsv1.StatefulSet{}
	err := h.GetClient().Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, depl)
	if err != nil {
		return depl, err
	}

	return depl, nil
}

// Delete - delete a statefulset.
func (s *StatefulSet) Delete(
	ctx context.Context,
	h *helper.Helper,
) error {
	err := h.GetClient().Delete(ctx, s.statefulset)
	if err != nil && !k8s_errors.IsNotFound(err) {
		err = fmt.Errorf("error deleting statefulset %s: %w", s.statefulset.Name, err)
		return err
	}

	return nil
}

// IsReady - validates when deployment is ready deployed to whats being requested
// - the requested replicas in the spec matches the ReadyReplicas of the status
// - both when the Generatation of the object matches the ObservedGeneration of the Status
func IsReady(deployment appsv1.StatefulSet) bool {
	return deployment.Spec.Replicas != nil &&
		*deployment.Spec.Replicas == deployment.Status.ReadyReplicas &&
		deployment.Generation == deployment.Status.ObservedGeneration
}
