package rbac

import (
	"context"
	"time"

	condition "github.com/openstack-k8s-operators/lib-common/modules/common/condition"
	helper "github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	common_role "github.com/openstack-k8s-operators/lib-common/modules/common/role"
	common_rolebinding "github.com/openstack-k8s-operators/lib-common/modules/common/rolebinding"
	common_serviceaccount "github.com/openstack-k8s-operators/lib-common/modules/common/serviceaccount"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// Reconciler - interface for rbac reconcilers
type Reconciler interface {
	// RbacConditionsSet - set the conditions on the instance
	RbacConditionsSet(c *condition.Condition)

	// RbacNamespace - return a string representing the namespace
	RbacNamespace() string

	// RbacResourceName - name of the resource to be used in the rbac resources (service, role, rolebinding))
	RbacResourceName() string
}

// ReconcileRbac - configures the serviceaccount, role, and role binding for the Reconciler instance
func ReconcileRbac(ctx context.Context, h *helper.Helper, instance Reconciler, rules []rbacv1.PolicyRule) (ctrl.Result, error) {

	// ServiceAccount
	sa := common_serviceaccount.NewServiceAccount(
		&corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      instance.RbacResourceName(),
				Namespace: instance.RbacNamespace(),
			},
		},
		time.Duration(10),
	)
	saResult, err := sa.CreateOrPatch(ctx, h)
	if err != nil {
		instance.RbacConditionsSet(condition.FalseCondition(
			condition.ServiceAccountReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			condition.ServiceAccountReadyErrorMessage,
			err.Error()))
		return saResult, err
	} else if (saResult != ctrl.Result{}) {
		instance.RbacConditionsSet(condition.FalseCondition(
			condition.ServiceAccountReadyCondition,
			condition.RequestedReason,
			condition.SeverityInfo,
			condition.ServiceAccountCreatingMessage))
		return saResult, nil
	}
	instance.RbacConditionsSet(condition.TrueCondition(
		condition.ServiceAccountReadyCondition,
		condition.ServiceAccountReadyMessage))

	// Role
	role := common_role.NewRole(
		&rbacv1.Role{
			ObjectMeta: metav1.ObjectMeta{
				Name:      instance.RbacResourceName() + "-role",
				Namespace: instance.RbacNamespace(),
			},
			Rules: rules,
		},
		time.Duration(10),
	)
	roleResult, err := role.CreateOrPatch(ctx, h)
	if err != nil {
		instance.RbacConditionsSet(condition.FalseCondition(
			condition.RoleReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			condition.RoleReadyErrorMessage,
			err.Error()))
		return roleResult, err
	} else if (roleResult != ctrl.Result{}) {
		instance.RbacConditionsSet(condition.FalseCondition(
			condition.RoleReadyCondition,
			condition.RequestedReason,
			condition.SeverityInfo,
			condition.RoleCreatingMessage))
		return roleResult, nil
	}
	instance.RbacConditionsSet(condition.TrueCondition(
		condition.RoleReadyCondition,
		condition.RoleReadyMessage))

	// RoleBinding
	rolebinding := common_rolebinding.NewRoleBinding(
		&rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      instance.RbacResourceName() + "-rolebinding",
				Namespace: instance.RbacNamespace(),
			},
			RoleRef: rbacv1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "Role",
				Name:     instance.RbacResourceName() + "-role",
			},
			Subjects: []rbacv1.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      instance.RbacResourceName(),
					Namespace: instance.RbacNamespace(),
				},
			},
		},
		time.Duration(10),
	)
	roleBindingResult, err := rolebinding.CreateOrPatch(ctx, h)
	if err != nil {
		instance.RbacConditionsSet(condition.FalseCondition(
			condition.RoleBindingReadyCondition,
			condition.ErrorReason,
			condition.SeverityWarning,
			condition.RoleBindingReadyErrorMessage,
			err.Error()))
		return roleBindingResult, err
	} else if (roleBindingResult != ctrl.Result{}) {
		instance.RbacConditionsSet(condition.FalseCondition(
			condition.RoleBindingReadyCondition,
			condition.RequestedReason,
			condition.SeverityInfo,
			condition.RoleBindingCreatingMessage))
		return roleBindingResult, nil
	}
	instance.RbacConditionsSet(condition.TrueCondition(
		condition.RoleReadyCondition,
		condition.RoleBindingReadyMessage))

	return ctrl.Result{}, nil
}
