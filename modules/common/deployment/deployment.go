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

package deployment

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	"github.com/openstack-k8s-operators/lib-common/modules/common/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// NewDeployment returns an initialized Deployment.
//
// Pass nil for pollingInterval and pollingTimeout to initialize with default.
func NewDeployment(
	deployment *appsv1.Deployment,
	timeout time.Duration,
) *Deployment {
	return &Deployment{
		deployment:          deployment,
		timeout:             timeout,
		rolloutPollInterval: ptr.To(DefaultPollInterval),
		rolloutPollTimeout:  ptr.To(DefaultPollTimeout),
	}
}

// SetRolloutPollInterval -
func (d *Deployment) SetRolloutPollInterval(interval time.Duration) {
	d.rolloutPollInterval = ptr.To(interval)
}

// GetRolloutPollInterval -
func (d *Deployment) GetRolloutPollInterval() *time.Duration {
	return d.rolloutPollInterval
}

// SetRolloutPollTimeout -
func (d *Deployment) SetRolloutPollTimeout(timeout time.Duration) {
	d.rolloutPollTimeout = ptr.To(timeout)
}

// GetRolloutPollTimeout -
func (d *Deployment) GetRolloutPollTimeout() *time.Duration {
	return d.rolloutPollTimeout
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
		if deployment.ObjectMeta.CreationTimestamp.IsZero() {
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
	// update the deployment object of the deployment type
	d.deployment = deployment

	h.GetLogger().Info(fmt.Sprintf("Deployment %s %s", deployment.Name, op))
	// Only poll on Deployment updates, not on initial create.
	if op != controllerutil.OperationResultCreated {
		// only poll if replicas > 0
		if d.deployment.Spec.Replicas != nil && *d.deployment.Spec.Replicas > 0 {
			// Ignore context.DeadlineExceeded when PollUntilContextTimeout reached
			// the poll timeout. d.rolloutStatus as information on the
			// replica rollout, the consumer can evaluate the rolloutStatus and
			// retry/reconcile until RolloutComplete, or ProgressDeadlineExceeded.
			if err := d.PollRolloutStatus(ctx, h); err != nil && !errors.Is(err, context.DeadlineExceeded) &&
				!strings.Contains(err.Error(), "would exceed context deadline") {
				return ctrl.Result{}, fmt.Errorf("poll rollout error: %w", err)
			}
		}
	}

	return ctrl.Result{}, nil
}

// PollRolloutStatus - will poll the deployment rollout to verify its status for Complet, Failed or polling until timeout.
//
// - Complete - all replicas updated using RolloutComplete()
//
// - Failed   - rollout of new config failed and the new pod is stuck in ProgressDeadlineExceeded using ProgressDeadlineExceeded()
func (d *Deployment) PollRolloutStatus(
	ctx context.Context,
	h *helper.Helper,
) error {
	if d.rolloutPollInterval == nil {
		d.rolloutPollInterval = ptr.To(DefaultPollInterval)
	}
	if d.rolloutPollTimeout == nil {
		d.rolloutPollTimeout = ptr.To(DefaultPollTimeout)
	}

	err := wait.PollUntilContextTimeout(ctx, *d.rolloutPollInterval, *d.rolloutPollTimeout, true, func(ctx context.Context) (bool, error) {
		// Fetch deployment object
		depl, err := GetDeploymentWithName(ctx, h, d.deployment.Name, d.deployment.Namespace)
		if err != nil {
			return false, err
		}
		d.deployment = depl

		// Check if rollout is complete
		if Complete(d.deployment.Status, d.deployment.Generation) {
			d.rolloutStatus = ptr.To(DeploymentPollCompleted)
			d.rolloutMessage = fmt.Sprintf(DeploymentPollCompletedMessage, d.deployment.Name)
			h.GetLogger().Info(d.rolloutMessage)
			// If rollout is complete, return true to stop polling
			return true, nil
		}

		// check if we already reached the ProgressDeadlineExceeded
		if ok, msg := ProgressDeadlineExceeded(d.deployment.Status); ok {
			d.rolloutStatus = ptr.To(DeploymentPollProgressDeadlineExceeded)
			d.rolloutMessage = fmt.Sprintf(DeploymentPollProgressDeadlineExceededMessage, d.deployment.Name, msg)
			// If rollout reached ProgressDeadlineExceeded, return true to stop polling
			return true, nil
		}

		// If not yet complete, continue waiting
		d.rolloutStatus = ptr.To(DeploymentPollProgressing)
		d.rolloutMessage = fmt.Sprintf(DeploymentPollProgressingMessage, d.deployment.Name,
			d.deployment.Status.UpdatedReplicas, d.deployment.Status.Replicas)
		h.GetLogger().Info(*d.rolloutStatus)

		return false, nil
	})

	return err
}

// RolloutComplete -
func (d *Deployment) RolloutComplete() bool {
	return d.GetRolloutStatus() != nil && *d.GetRolloutStatus() == DeploymentPollCompleted
}

// Complete -
func Complete(status appsv1.DeploymentStatus, generation int64) bool {
	return status.UpdatedReplicas == status.Replicas &&
		status.Replicas == status.AvailableReplicas &&
		status.ObservedGeneration == generation
}

// ProgressDeadlineExceeded -
func ProgressDeadlineExceeded(status appsv1.DeploymentStatus) (bool, string) {
	for _, condition := range status.Conditions {
		if condition.Type == appsv1.DeploymentProgressing &&
			condition.Status == corev1.ConditionFalse &&
			condition.Reason == DeploymentPollProgressDeadlineExceeded {
			return true, condition.Message
		}
	}

	return false, ""
}

// GetRolloutStatus - get rollout status of the deployment.
func (d *Deployment) GetRolloutStatus() *string {
	return d.rolloutStatus
}

// GetRolloutMessage - get rollout message of the deployment.
func (d *Deployment) GetRolloutMessage() string {
	return d.rolloutMessage
}

// Delete - delete a deployment.
func (d *Deployment) Delete(
	ctx context.Context,
	h *helper.Helper,
) error {
	err := h.GetClient().Delete(ctx, d.deployment)
	if err != nil && !k8s_errors.IsNotFound(err) {
		return fmt.Errorf("Error deleting deployment %s: %w", d.deployment.Name, err)
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
