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

package database

import (
	"context"
	"fmt"
	"time"

	"github.com/openstack-k8s-operators/lib-common/pkg/common"
	"github.com/openstack-k8s-operators/lib-common/pkg/condition"
	"github.com/openstack-k8s-operators/lib-common/pkg/helper"
	mariadbv1 "github.com/openstack-k8s-operators/mariadb-operator/api/v1beta1"

	corev1 "k8s.io/api/core/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// NewDatabase returns an initialized DB.
func NewDatabase(
	databaseName string,
	databaseUser string,
	secret string,
	labels map[string]string,
) *Database {
	return &Database{
		databaseName: databaseName,
		databaseUser: databaseUser,
		secret:       secret,
		labels:       labels,
	}
}

//
// setDatabaseHostname - set the service name of the DB as the databaseHostname
//
func (d *Database) setDatabaseHostname(
	ctx context.Context,
	h *helper.Helper,
) (condition.Condition, error) {

	selector := map[string]string{
		"app": "mariadb",
	}
	serviceList, err := common.GetServicesListWithLabel(
		ctx,
		h,
		h.GetBeforeObject().GetNamespace(),
		selector,
	)
	if err != nil || len(serviceList.Items) == 0 {
		msg := fmt.Sprintf("Error getting the DB service using label %v", selector)
		cond := condition.NewCondition(
			condition.TypeError,
			corev1.ConditionTrue,
			ReasonDBServiceNameError,
			msg)

		return cond, err
	}

	// can we expect there is only one DB instance per namespace?
	if len(serviceList.Items) > 1 {
		msg := fmt.Sprintf("more then one DB service found %d", len(serviceList.Items))
		cond := condition.NewCondition(
			condition.TypeError,
			corev1.ConditionTrue,
			ReasonDBServiceNameError,
			msg)

		return cond, err
	}
	d.databaseHostname = serviceList.Items[0].GetName()

	return condition.Condition{}, nil
}

//
// GetDatabaseHostname - returns the DB hostname which host the DB
//
func (d *Database) GetDatabaseHostname() string {
	return d.databaseHostname
}

//
// GetDatabase - returns the DB
//
func (d *Database) GetDatabase() *mariadbv1.MariaDBDatabase {
	return d.database
}

//
// CreateOrPatchDB - create or patch the service DB instance
//
func (d *Database) CreateOrPatchDB(
	ctx context.Context,
	h *helper.Helper,
) (condition.Condition, ctrl.Result, error) {

	db := &mariadbv1.MariaDBDatabase{
		ObjectMeta: metav1.ObjectMeta{
			Name:      h.GetBeforeObject().GetName(),
			Namespace: h.GetBeforeObject().GetNamespace(),
		},
		Spec: mariadbv1.MariaDBDatabaseSpec{
			// the DB name must not change, therefore specify it outside the mutuate function
			Name: d.databaseName,
		},
	}

	// set the database hostname on the db instance
	cond, err := d.setDatabaseHostname(ctx, h)
	if err != nil {
		return cond, ctrl.Result{}, err
	}

	op, err := controllerutil.CreateOrPatch(ctx, h.GetClient(), db, func() error {
		db.Labels = common.MergeStringMaps(
			db.GetLabels(),
			d.labels,
		)

		db.Spec.Secret = d.secret

		err := controllerutil.SetControllerReference(h.GetBeforeObject(), db, h.GetScheme())
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil && !k8s_errors.IsNotFound(err) {
		msg := fmt.Sprintf("Error create or update DB object %s", db.Name)
		c := condition.NewCondition(
			condition.TypeError,
			corev1.ConditionTrue,
			ReasonDBPatchError,
			msg)

		return c, ctrl.Result{}, fmt.Errorf("%s - %s", msg, err.Error())
	}

	if op != controllerutil.OperationResultNone {
		msg := fmt.Sprintf("DB object %s created or patched", db.Name)
		c := condition.NewCondition(
			condition.TypeCreated,
			corev1.ConditionTrue,
			ReasonDBPatchOK,
			msg,
		)
		return c, ctrl.Result{RequeueAfter: time.Second * 5}, nil
	}

	cond, err = d.getDBWithName(
		ctx,
		h,
	)
	if err != nil {
		return cond, ctrl.Result{}, err
	}

	return condition.Condition{}, ctrl.Result{}, nil
}

//
// WaitForDBCreated - wait until the MariaDBDatabase is initialized and reports Status.Completed == true
//
func (d *Database) WaitForDBCreated(
	ctx context.Context,
	h *helper.Helper,
) (condition.Condition, ctrl.Result, error) {

	cond, err := d.getDBWithName(
		ctx,
		h,
	)
	if err != nil && !k8s_errors.IsNotFound(err) {
		return cond, ctrl.Result{}, err
	}

	if !d.database.Status.Completed || k8s_errors.IsNotFound(err) {
		msg := fmt.Sprintf("Waiting for service DB %s to be created", d.database.Name)
		cond := condition.NewCondition(
			condition.TypeWaiting,
			corev1.ConditionTrue,
			ReasonDBWaitingInitialized,
			msg)
		return cond, ctrl.Result{RequeueAfter: time.Second * 5}, nil
	}

	return condition.Condition{}, ctrl.Result{}, nil
}

//
// getDBWithName - get DB object with name in namespace
//
func (d *Database) getDBWithName(
	ctx context.Context,
	h *helper.Helper,
) (condition.Condition, error) {
	db := &mariadbv1.MariaDBDatabase{}
	err := h.GetClient().Get(
		ctx,
		types.NamespacedName{
			Name:      d.databaseName,
			Namespace: h.GetBeforeObject().GetNamespace(),
		},
		db)
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			msg := fmt.Sprintf("Failed to get %s database %s ", d.databaseName, h.GetBeforeObject().GetNamespace())
			cond := condition.NewCondition(
				condition.TypeError,
				corev1.ConditionTrue,
				ReasonDBNotFound,
				msg)

			return cond, common.WrapErrorForObject(msg, h.GetBeforeObject(), err)
		}

		msg := fmt.Sprintf("DB error %s %s ", d.databaseName, h.GetBeforeObject().GetNamespace())
		cond := condition.NewCondition(
			condition.TypeError,
			corev1.ConditionTrue,
			ReasonDBError,
			msg)

		return cond, common.WrapErrorForObject(msg, h.GetBeforeObject(), err)
	}

	d.database = db

	return condition.Condition{}, nil
}
