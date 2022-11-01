/*
Copyright 2022 Red Hat

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

	"github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	"github.com/openstack-k8s-operators/lib-common/modules/common/service"
	"github.com/openstack-k8s-operators/lib-common/modules/common/util"
	mariadbv1 "github.com/openstack-k8s-operators/mariadb-operator/api/v1beta1"

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
) error {

	selector := map[string]string{
		"app": "mariadb",
	}
	serviceList, err := service.GetServicesListWithLabel(
		ctx,
		h,
		h.GetBeforeObject().GetNamespace(),
		selector,
	)
	if err != nil || len(serviceList.Items) == 0 {
		return fmt.Errorf("Error getting the DB service using label %v: %w",
			selector, err)
	}

	// can we expect there is only one DB instance per namespace?
	if len(serviceList.Items) > 1 {
		return util.WrapErrorForObject(
			fmt.Sprintf("more then one DB service found %d", len(serviceList.Items)),
			d.database,
			err,
		)
	}
	d.databaseHostname = serviceList.Items[0].GetName()

	return nil
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
) (ctrl.Result, error) {

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
	err := d.setDatabaseHostname(ctx, h)
	if err != nil {
		return ctrl.Result{}, err
	}

	op, err := controllerutil.CreateOrPatch(ctx, h.GetClient(), db, func() error {
		db.Labels = util.MergeStringMaps(
			db.GetLabels(),
			d.labels,
		)

		db.Spec.Secret = d.secret

		err := controllerutil.SetControllerReference(h.GetBeforeObject(), db, h.GetScheme())
		if err != nil {
			return err
		}

		// If the service object doesn't have our finalizer, add it.
		controllerutil.AddFinalizer(db, h.GetFinalizer())

		return nil
	})

	if err != nil && !k8s_errors.IsNotFound(err) {
		return ctrl.Result{}, util.WrapErrorForObject(
			fmt.Sprintf("Error create or update DB object %s", db.Name),
			db,
			err,
		)
	}

	if op != controllerutil.OperationResultNone {
		return ctrl.Result{RequeueAfter: time.Second * 5}, util.WrapErrorForObject(
			fmt.Sprintf("DB object %s created or patched", db.Name),
			db,
			err,
		)
	}

	err = d.getDBWithName(
		ctx,
		h,
	)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

//
// WaitForDBCreated - wait until the MariaDBDatabase is initialized and reports Status.Completed == true
//
func (d *Database) WaitForDBCreated(
	ctx context.Context,
	h *helper.Helper,
) (ctrl.Result, error) {

	err := d.getDBWithName(
		ctx,
		h,
	)
	if err != nil && !k8s_errors.IsNotFound(err) {
		return ctrl.Result{}, err
	}

	if !d.database.Status.Completed || k8s_errors.IsNotFound(err) {
		util.LogForObject(
			h,
			fmt.Sprintf("Waiting for service DB %s to be created", d.database.Name),
			d.database,
		)

		return ctrl.Result{RequeueAfter: time.Second * 5}, nil
	}

	return ctrl.Result{}, nil
}

//
// getDBWithName - get DB object with name in namespace
//
func (d *Database) getDBWithName(
	ctx context.Context,
	h *helper.Helper,
) error {
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
			return util.WrapErrorForObject(
				fmt.Sprintf("Failed to get %s database %s ", d.databaseName, h.GetBeforeObject().GetNamespace()),
				h.GetBeforeObject(),
				err,
			)
		}

		return util.WrapErrorForObject(
			fmt.Sprintf("DB error %s %s ", d.databaseName, h.GetBeforeObject().GetNamespace()),
			h.GetBeforeObject(),
			err,
		)
	}

	d.database = db

	return nil
}

//
// GetDatabaseByName returns a *Database object with specified name and namespace
//
func GetDatabaseByName(
	ctx context.Context,
	h *helper.Helper,
	name string,
) (*Database, error) {
	// create a Database by suppplying a resource name
	db := &Database{
		databaseName: name,
	}
	// then querying the MariaDBDatabase and store it in db by calling
	if err := db.getDBWithName(ctx, h); err != nil {
		return db, err
	}
	return db, nil
}

//
// DeleteFinalizer deletes a finalizer by its object
//
func (d *Database) DeleteFinalizer(
	ctx context.Context,
	h *helper.Helper,
) error {
	controllerutil.RemoveFinalizer(d.database, h.GetFinalizer())
	if err := h.GetClient().Update(ctx, d.database); err != nil && !k8s_errors.IsNotFound(err) {
		return err
	}
	return nil
}
