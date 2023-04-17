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

package rolebinding

import (
	"context"
	"fmt"
	"time"

	"github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	"github.com/openstack-k8s-operators/lib-common/modules/common/util"
	rbacv1 "k8s.io/api/rbac/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// NewRoleBinding returns an initialized RoleBinding
func NewRoleBinding(
	roleBinding *rbacv1.RoleBinding,
	labels map[string]string,
	timeout time.Duration,
) *RoleBinding {
	return &RoleBinding{
		roleBinding: roleBinding,
		timeout:        timeout,
	}
}

// CreateOrPatch - creates or patches a route, reconciles after Xs if object won't exist.
func (r *RoleBinding) CreateOrPatch(
	ctx context.Context,
	h *helper.Helper,
) (ctrl.Result, error) {
	rb := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.roleBinding.Name,
			Namespace: r.roleBinding.Namespace,
		},
	}

	op, err := controllerutil.CreateOrPatch(ctx, h.GetClient(), rb, func() error {
		rb.Labels = util.MergeStringMaps(rb.Labels, r.roleBinding.Labels)
		rb.Annotations = r.roleBinding.Annotations

		err := controllerutil.SetControllerReference(h.GetBeforeObject(), rb, h.GetScheme())
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			h.GetLogger().Info(fmt.Sprintf("RoleBinding %s not found, reconcile in %s", rb.Name, r.timeout))
			return ctrl.Result{RequeueAfter: r.timeout}, nil
		}
		return ctrl.Result{}, err
	}
	if op != controllerutil.OperationResultNone {
		h.GetLogger().Info(fmt.Sprintf("RoleBinding %s - %s", rb.Name, op))
	}

	return ctrl.Result{}, nil
}

// Delete - delete a RoleBinding
func (r *RoleBinding) Delete(
	ctx context.Context,
	h *helper.Helper,
) error {

	err := h.GetClient().Delete(ctx, r.roleBinding)
	if err != nil && !k8s_errors.IsNotFound(err) {
		err = fmt.Errorf("Error deleting roleBinding %s: %v", r.roleBinding.Name, err)
		return err
	}

	return nil
}
