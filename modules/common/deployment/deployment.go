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

// Package deployment provides utilities for managing Kubernetes Deployment resources
package deployment

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

// NewDeployment returns an initialized Deployment.
func NewDeployment(
	deployment *appsv1.Deployment,
	timeout time.Duration,
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
		// Deployment selector is immutable so we set this value only if
		// a new object is going to be created
		if deployment.CreationTimestamp.IsZero() {
			deployment.Spec.Selector = d.deployment.Spec.Selector
		}
		deployment.Annotations = util.MergeStringMaps(deployment.Annotations, d.deployment.Annotations)
		deployment.Labels = util.MergeStringMaps(deployment.Labels, d.deployment.Labels)
		deployment.Spec.Template = d.deployment.Spec.Template
		deployment.Spec.Replicas = d.deployment.Spec.Replicas
		deployment.Spec.Strategy = d.deployment.Spec.Strategy

		err := controllerutil.SetControllerReference(h.GetBeforeObject(), deployment, h.GetScheme())
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			h.GetLogger().Info(fmt.Sprintf("Deployment %s not found, reconcile in %s", deployment.Name, d.timeout))
			return ctrl.Result{RequeueAfter: d.timeout}, nil
		}
		return ctrl.Result{}, err
	}
	if op != controllerutil.OperationResultNone {
		h.GetLogger().Info(fmt.Sprintf("Deployment %s - %s", deployment.Name, op))
	}

	// update the deployment object of the deployment type
	d.deployment, err = GetDeploymentWithName(ctx, h, deployment.GetName(), deployment.GetNamespace())
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// Delete - delete a deployment.
func (d *Deployment) Delete(
	ctx context.Context,
	h *helper.Helper,
) error {
	err := h.GetClient().Delete(ctx, d.deployment)
	if err != nil && !k8s_errors.IsNotFound(err) {
		return fmt.Errorf("error deleting deployment %s: %w", d.deployment.Name, err)
	}

	return nil
}

// GetDeployment - get the deployment object.
func (d *Deployment) GetDeployment() appsv1.Deployment {
	return *d.deployment
}

// GetDeploymentWithName func
func GetDeploymentWithName(
	ctx context.Context,
	h *helper.Helper,
	name string,
	namespace string,
) (*appsv1.Deployment, error) {

	depl := &appsv1.Deployment{}
	err := h.GetClient().Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, depl)
	if err != nil {
		return depl, err
	}

	return depl, nil
}

// IsReady - validates when deployment is ready deployed to whats being requested
// - the requested replicas in the spec matches the ReadyReplicas of the status
// - the Status.Replicas match Status.ReadyReplicas. if a deployment update is in progress, Replicas > ReadyReplicas
// - both when the Generatation of the object matches the ObservedGeneration of the Status
func IsReady(deployment appsv1.Deployment) bool {
	return deployment.Spec.Replicas != nil &&
		*deployment.Spec.Replicas == deployment.Status.ReadyReplicas &&
		deployment.Status.Replicas == deployment.Status.ReadyReplicas &&
		deployment.Generation == deployment.Status.ObservedGeneration
}
