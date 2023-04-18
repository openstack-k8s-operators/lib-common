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

package serviceaccount

import (
	"context"
	"fmt"
	"time"

	"github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	"github.com/openstack-k8s-operators/lib-common/modules/common/util"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// NewServiceAccount returns an initialized ServiceAccount
func NewServiceAccount(
	serviceAccount *corev1.ServiceAccount,
	labels map[string]string,
	timeout time.Duration,
) *ServiceAccount {
	return &ServiceAccount{
		serviceAccount: serviceAccount,
		timeout:        timeout,
	}
}

// CreateOrPatch - creates or patches a service account, reconciles after Xs if object won't exist.
func (s *ServiceAccount) CreateOrPatch(
	ctx context.Context,
	h *helper.Helper,
) (ctrl.Result, error) {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.serviceAccount.Name,
			Namespace: s.serviceAccount.Namespace,
		},
	}

	op, err := controllerutil.CreateOrPatch(ctx, h.GetClient(), sa, func() error {
		sa.Labels = util.MergeStringMaps(sa.Labels, s.serviceAccount.Labels)
		sa.Annotations = util.MergeStringMaps(sa.Labels, s.serviceAccount.Annotations)

		err := controllerutil.SetControllerReference(h.GetBeforeObject(), sa, h.GetScheme())
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			h.GetLogger().Info(fmt.Sprintf("ServiceAccount %s not found, reconcile in %s", sa.Name, s.timeout))
			return ctrl.Result{RequeueAfter: s.timeout}, nil
		}
		return ctrl.Result{}, err
	}
	if op != controllerutil.OperationResultNone {
		h.GetLogger().Info(fmt.Sprintf("ServiceAccount %s - %s", sa.Name, op))
	}

	return ctrl.Result{}, nil
}

// Delete - delete a service.
func (s *ServiceAccount) Delete(
	ctx context.Context,
	h *helper.Helper,
) error {

	err := h.GetClient().Delete(ctx, s.serviceAccount)
	if err != nil && !k8s_errors.IsNotFound(err) {
		err = fmt.Errorf("Error deleting serviceAccount %s: %w", s.serviceAccount.Name, err)
		return err
	}

	return nil
}
