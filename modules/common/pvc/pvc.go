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

package pvc

import (
	"context"
	"fmt"
	"time"

	"github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	"github.com/openstack-k8s-operators/lib-common/modules/common/util"
	corev1 "k8s.io/api/core/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// NewPvc returns an initialized Pvc.
func NewPvc(
	pvc *corev1.PersistentVolumeClaim,
	timeout time.Duration,
) *Pvc {
	return &Pvc{
		pvc:     pvc,
		timeout: timeout,
	}
}

// CreateOrPatch - creates or patches a pvc, reconciles after Xs if object won't exist.
func (p *Pvc) CreateOrPatch(
	ctx context.Context,
	h *helper.Helper,
) (ctrl.Result, error) {
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      p.pvc.Name,
			Namespace: p.pvc.Namespace,
		},
	}

	op, err := controllerutil.CreateOrPatch(ctx, h.GetClient(), pvc, func() error {
		pvc.Annotations = util.MergeStringMaps(pvc.Annotations, p.pvc.Annotations)
		pvc.Labels = util.MergeStringMaps(pvc.Labels, p.pvc.Labels)

		// For now, we don't support changes to existing PVC specs for the
		// following fields.  Technically it is possible to change the size
		// request, but this requires dynamic provisioning and a storage
		// class that supports such a thing.
		if pvc.ObjectMeta.CreationTimestamp.IsZero() {
			pvc.Spec.Resources.Requests = p.pvc.Spec.Resources.Requests
			pvc.Spec.StorageClassName = p.pvc.Spec.StorageClassName
			pvc.Spec.AccessModes = p.pvc.Spec.AccessModes
		}

		err := controllerutil.SetControllerReference(h.GetBeforeObject(), pvc, h.GetScheme())

		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if k8s_errors.IsNotFound(err) {
			h.GetLogger().Info(fmt.Sprintf("Pvc %s not found, reconcile in %s", pvc.Name, p.timeout))
			return ctrl.Result{RequeueAfter: p.timeout}, nil
		}
		h.GetRecorder().Event(h.GetBeforeObject(), corev1.EventTypeWarning, "PvcError", fmt.Sprintf("error create/updating pvc: %s", p.pvc.Name))
		return ctrl.Result{}, err
	}
	if op == controllerutil.OperationResultCreated {
		h.GetRecorder().Event(h.GetBeforeObject(), corev1.EventTypeNormal, "PvcCreated", fmt.Sprintf("pvc %s created", p.pvc.Name))
	}
	if op != controllerutil.OperationResultNone {
		h.GetLogger().Info(fmt.Sprintf("Pvc %s - %s", pvc.Name, op))
	}

	// update the pvc object of the pvc type
	p.pvc, err = GetPvcWithName(ctx, h, pvc.GetName(), pvc.GetNamespace())

	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// GetPvc - get the pvc object.
func (p *Pvc) GetPvc() corev1.PersistentVolumeClaim {
	return *p.pvc
}

// GetPvcWithName func
func GetPvcWithName(
	ctx context.Context,
	h *helper.Helper,
	name string,
	namespace string,
) (*corev1.PersistentVolumeClaim, error) {

	pvc := &corev1.PersistentVolumeClaim{}
	err := h.GetClient().Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, pvc)
	if err != nil {
		return pvc, err
	}

	return pvc, nil
}
