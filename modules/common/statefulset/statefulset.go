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
	timeout int,
) *StatefulSet {
	return &StatefulSet{
		statefulset: statefulset,
		timeout:     time.Duration(timeout) * time.Second,
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
		// selector is immutable so we set this value only if
		// a new object is going to be created
		if statefulset.ObjectMeta.CreationTimestamp.IsZero() {
			statefulset.Spec.Selector = s.statefulset.Spec.Selector
		}

		statefulset.Annotations = util.MergeStringMaps(statefulset.Annotations, s.statefulset.Annotations)
		statefulset.Labels = util.MergeStringMaps(statefulset.Labels, s.statefulset.Labels)
		// We need to copy the Spec field by field as Selector is not updatable
		// This list needs to be synced StatefulSet to gain ability to set
		// those new fields via lib-common
		statefulset.Spec.Replicas = s.statefulset.Spec.Replicas
		statefulset.Spec.Template = s.statefulset.Spec.Template
		statefulset.Spec.VolumeClaimTemplates = s.statefulset.Spec.VolumeClaimTemplates
		statefulset.Spec.ServiceName = s.statefulset.Spec.ServiceName
		statefulset.Spec.PodManagementPolicy = s.statefulset.Spec.PodManagementPolicy
		statefulset.Spec.UpdateStrategy = s.statefulset.Spec.UpdateStrategy
		statefulset.Spec.RevisionHistoryLimit = s.statefulset.Spec.RevisionHistoryLimit
		statefulset.Spec.MinReadySeconds = s.statefulset.Spec.MinReadySeconds
		statefulset.Spec.PersistentVolumeClaimRetentionPolicy = s.statefulset.Spec.PersistentVolumeClaimRetentionPolicy

		err := controllerutil.SetControllerReference(h.GetBeforeObject(), statefulset, h.GetScheme())
		if err != nil {
			return err
		}

		return nil
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
		err = fmt.Errorf("Error deleting statefulset %s: %v", s.statefulset.Name, err)
		return err
	}

	return nil
}

// SetTimeout defines the duration used for requeueing while waiting for the
// stateful set to exist.
func (s *StatefulSet) SetTimeout(timeout time.Duration) {
	s.timeout = timeout
}
