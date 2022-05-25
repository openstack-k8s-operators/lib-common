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
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/openstack-k8s-operators/lib-common/pkg/helper"

	keystonev1 "github.com/openstack-k8s-operators/keystone-operator/api/v1beta1"
	mariadbv1 "github.com/openstack-k8s-operators/mariadb-operator/api/v1beta1"
)

var trueVal = true
var dbObj = &mariadbv1.MariaDBDatabase{
	TypeMeta: metav1.TypeMeta{
		Kind:       "MariaDBDatabase",
		APIVersion: "mariadb.openstack.org/v1beta1",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:            "keystone",
		Namespace:       "openstack",
		ResourceVersion: "1",
		Labels: map[string]string{
			"label-key": "label-value",
		},
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion:         "keystone.openstack.org/v1beta1",
				Kind:               "KeystoneAPI",
				Name:               "keystone",
				UID:                "",
				Controller:         &trueVal,
				BlockOwnerDeletion: &trueVal,
			},
		},
	},
	Spec: mariadbv1.MariaDBDatabaseSpec{
		Secret: "dbsecret",
		Name:   "keystone",
	},
}

var keystoneObj = &keystonev1.KeystoneAPI{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "keystone",
		Namespace: "openstack",
	},
}

func TestCreateOrPatchDB(t *testing.T) {
	t.Run("Create database", func(t *testing.T) {
		g := NewWithT(t)

		scheme := runtime.NewScheme()
		_ = mariadbv1.AddToScheme(scheme)
		_ = keystonev1.AddToScheme(scheme)

		clientBuilder := fake.NewClientBuilder().WithScheme(scheme).Build()

		d := NewDatabase(
			"keystone",
			"dbuser",
			"dbsecret",
			map[string]string{"label-key": "label-value"},
		)

		kclient, _ := kubernetes.NewForConfig(&rest.Config{})
		h, err := helper.NewHelper(
			keystoneObj,
			clientBuilder,
			kclient,
			scheme,
			ctrl.Log.WithName("test").WithName("test"),
		)
		if err != nil {
			t.Fatalf("NewHelper error: (%v)", err)
		}

		// createDB
		_, _, err = d.CreateOrPatchDB(
			context.TODO(),
			h,
		)
		if err != nil {
			t.Fatalf("CreateOrPatchDB error: (%v)", err)
		}

		db, _, err := d.GetDBWithName(context.TODO(), h)
		if err != nil {
			t.Fatalf("GetDBWithName error: (%v) -%v", err, db)
		}
		g.Expect(db).To(Equal(dbObj))
	})
}

func TestGetDBWithName(t *testing.T) {
	t.Run("Get database with name", func(t *testing.T) {
		g := NewWithT(t)

		scheme := runtime.NewScheme()
		_ = mariadbv1.AddToScheme(scheme)
		_ = keystonev1.AddToScheme(scheme)

		// Objects to track in the fake client.
		Objs := []runtime.Object{}
		Objs = append(Objs, dbObj)

		// add Objs to the cache
		clientBuilder := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(Objs...).Build()

		kclient, _ := kubernetes.NewForConfig(&rest.Config{})

		h, err := helper.NewHelper(
			keystoneObj,
			clientBuilder,
			kclient,
			scheme,
			ctrl.Log.WithName("test").WithName("test"),
		)
		if err != nil {
			t.Fatalf("NewHelper error: (%v)", err)
		}

		d := NewDatabase(
			"keystone",
			"dbuser",
			"dbsecret",
			map[string]string{"label-key": "label-value"},
		)

		db, _, err := d.GetDBWithName(context.TODO(), h)
		if err != nil {
			t.Fatalf("GetDBWithName error: (%v) -%v", err, db)
		}
		g.Expect(db.Spec.Name).To(Equal("keystone"))
		g.Expect(db.Spec.Secret).To(Equal("dbsecret"))
	})
}
