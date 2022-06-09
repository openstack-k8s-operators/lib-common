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

package common

import (
	"context"
	"fmt"
	"time"

	"github.com/openstack-k8s-operators/lib-common/pkg/helper"
	appsv1 "k8s.io/api/apps/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// NewDeployment returns an initialized Deployment.
func NewDeployment(
	deployment *appsv1.Deployment,
	timeout int,
) *Deployment {
	return &Deployment{
		deployment: deployment,
		timeout:    timeout,
	}
}

// CreateOrPatch - creates or patches a deployment, reconciles after Xs if object won't exist.
func (d *Deployment) CreateOrPatch(
	ctx context.Context,
	h *helper.Helper,
) (ctrl.Result, error) {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      d.deployment.Name,
			Namespace: d.deployment.Namespace,
		},
	}

	op, err := controllerutil.CreateOrPatch(ctx, h.GetClient(), deployment, func() error {
		deployment.Annotations = d.deployment.Annotations
		deployment.Labels = MergeStringMaps(deployment.Labels, d.deployment.Labels)
		deployment.Spec = d.deployment.Spec

		err := controllerutil.SetControllerReference(h.GetBeforeObject(), deployment, h.GetScheme())
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			h.GetLogger().Info(fmt.Sprintf("Deployment %s not found, reconcile in %ds", deployment.Name, d.timeout))
			return ctrl.Result{RequeueAfter: time.Duration(d.timeout) * time.Second}, nil
		}
		return ctrl.Result{}, err
	}
	if op != controllerutil.OperationResultNone {
		h.GetLogger().Info(fmt.Sprintf("Deployment %s - %s", deployment.Name, op))
	}

	return ctrl.Result{}, nil
}

// Delete - delete a seployment.
func (d *Deployment) Delete(
	ctx context.Context,
	h *helper.Helper,
) error {
	err := h.GetClient().Delete(ctx, d.deployment)
	if err != nil && !k8s_errors.IsNotFound(err) {
		err = fmt.Errorf("Error deleting deployment %s: %v", d.deployment.Name, err)
		return err
	}

	return nil
}
