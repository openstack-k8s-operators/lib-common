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
// by looking up the Service via the name of the MariaDB CR which provides it.
//
func (d *Database) setDatabaseHostname(
	ctx context.Context,
	h *helper.Helper,
	name string,
) error {

	// When the MariaDB CR provides the Service it sets the "cr" label of the
	// Service to "mariadb-<name of the MariaDB CR>". So we use this label
	// to select the right Service. See:
	// https://github.com/openstack-k8s-operators/mariadb-operator/blob/5781b0cf1087d7d28fa285bd5c44689acba92183/pkg/service.go#L17
	// https://github.com/openstack-k8s-operators/mariadb-operator/blob/590ffdc5ad86fe653f9cd8a7102bb76dfe2e36d1/pkg/utils.go#L4
	selector := map[string]string{
		"app": "mariadb",
		"cr":  fmt.Sprintf("mariadb-%s", name),
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

	// We assume here that a MariaDB CR instance always creates a single
	// Service. If multiple DB services are used the they are managed via
	// separate MariaDB CRs.
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
// Deprecated. Use CreateOrPatchDBByName instead. If you want to use the
// default the DB service instance of the deployment then pass "openstack" as
// the name.
//
func (d *Database) CreateOrPatchDB(
	ctx context.Context,
	h *helper.Helper,
) (ctrl.Result, error) {
	return d.CreateOrPatchDBByName(ctx, h, "openstack")
}

//
// CreateOrPatchDBByName - create or patch the service DB instance on
// the DB service. The DB service is selected by the name of the MariaDB CR
// providing the service.
//
func (d *Database) CreateOrPatchDBByName(
	ctx context.Context,
	h *helper.Helper,
	name string,
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
	err := d.setDatabaseHostname(ctx, h, name)
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
