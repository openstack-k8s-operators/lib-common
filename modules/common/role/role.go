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

// Package role provides utilities for managing Kubernetes Role and RoleBinding resources
package role

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

// NewRole returns an initialized Role
func NewRole(
	role *rbacv1.Role,
	timeout time.Duration,
) *Role {
	return &Role{
		role:    role,
		timeout: timeout,
	}
}

// CreateOrPatch - creates or patches a role, reconciles after Xs if object won't exist.
func (r *Role) CreateOrPatch(
	ctx context.Context,
	h *helper.Helper,
) (ctrl.Result, error) {
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.role.Name,
			Namespace: r.role.Namespace,
		},
	}

	op, err := controllerutil.CreateOrPatch(ctx, h.GetClient(), role, func() error {
		role.Labels = util.MergeStringMaps(role.Labels, r.role.Labels)
		role.Annotations = util.MergeStringMaps(role.Labels, r.role.Annotations)
		role.Rules = r.role.Rules
		err := controllerutil.SetControllerReference(h.GetBeforeObject(), role, h.GetScheme())
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			h.GetLogger().Info(fmt.Sprintf("Role %s not found, reconcile in %s", role.Name, r.timeout))
			return ctrl.Result{RequeueAfter: r.timeout}, nil
		}
		return ctrl.Result{}, util.WrapErrorForObject(
			fmt.Sprintf("Error creating role %s", role.Name),
			role,
			err,
		)
	}
	if op != controllerutil.OperationResultNone {
		h.GetLogger().Info(fmt.Sprintf("Role %s - %s", role.Name, op))
	}

	return ctrl.Result{}, nil
}

// Delete - delete a role
func (r *Role) Delete(
	ctx context.Context,
	h *helper.Helper,
) error {

	err := h.GetClient().Delete(ctx, r.role)
	if err != nil && !k8s_errors.IsNotFound(err) {
		err = fmt.Errorf("error deleting role %s: %w", r.role.Name, err)
		return err
	}

	return nil
}
