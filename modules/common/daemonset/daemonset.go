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

// Package daemonset provides utilities for managing Kubernetes DaemonSet resources
package daemonset

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

// NewDaemonSet returns an initialized DaemonSet
func NewDaemonSet(
	daemonset *appsv1.DaemonSet,
	timeout time.Duration,
) *DaemonSet {
	return &DaemonSet{
		daemonset: daemonset,
		timeout:   timeout,
	}
}

// CreateOrPatch - creates or patches a DaemonSet, reconciles after Xs if object won't exist.
func (d *DaemonSet) CreateOrPatch(
	ctx context.Context,
	h *helper.Helper,
) (ctrl.Result, error) {
	daemonset := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      d.daemonset.Name,
			Namespace: d.daemonset.Namespace,
		},
	}

	op, err := controllerutil.CreateOrPatch(ctx, h.GetClient(), daemonset, func() error {
		// DaemonSet selector is immutable so we set this value only if
		// a new object is going to be created
		if daemonset.CreationTimestamp.IsZero() {
			daemonset.Spec.Selector = d.daemonset.Spec.Selector
		}
		daemonset.Annotations = util.MergeStringMaps(daemonset.Annotations, d.daemonset.Annotations)
		daemonset.Labels = util.MergeStringMaps(daemonset.Labels, d.daemonset.Labels)
		daemonset.Spec.Template = d.daemonset.Spec.Template
		daemonset.Spec.UpdateStrategy = d.daemonset.Spec.UpdateStrategy

		err := controllerutil.SetControllerReference(h.GetBeforeObject(), daemonset, h.GetScheme())
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			util.LogForObject(h, fmt.Sprintf("DaemonSet not found, reconcile in %s", d.timeout), daemonset)
			return ctrl.Result{RequeueAfter: d.timeout}, nil
		}
		return ctrl.Result{}, err
	}
	if op != controllerutil.OperationResultNone {
		util.LogForObject(h, fmt.Sprintf("DaemonSet: %s", op), daemonset)
	}

	// update the daemonset object of the daemonset type
	d.daemonset, err = GetDaemonSetWithName(ctx, h, daemonset.GetName(), daemonset.GetNamespace())
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			return ctrl.Result{RequeueAfter: d.timeout}, nil
		}
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// Delete - delete a daemonset.
func (d *DaemonSet) Delete(
	ctx context.Context,
	h *helper.Helper,
) error {
	err := h.GetClient().Delete(ctx, d.daemonset)
	if err != nil && !k8s_errors.IsNotFound(err) {
		return fmt.Errorf("error deleting daemonset %s: %w", d.daemonset.Name, err)
	}

	return nil
}

// GetDaemonSet - get the daemonset object.
func (d *DaemonSet) GetDaemonSet() appsv1.DaemonSet {
	return *d.daemonset
}

// GetDaemonSetWithName - get the daemonset object with a given name.
func GetDaemonSetWithName(
	ctx context.Context,
	h *helper.Helper,
	name string,
	namespace string,
) (*appsv1.DaemonSet, error) {

	dset := &appsv1.DaemonSet{}
	err := h.GetClient().Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, dset)
	if err != nil {
		return dset, err
	}

	return dset, nil
}
