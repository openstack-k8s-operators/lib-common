/*
Copyright 2026 Red Hat

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

package configmap

import (
	"context"
	"errors"
	"testing"

	"github.com/go-logr/logr"
	"github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	"github.com/openstack-k8s-operators/lib-common/modules/common/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	kubernetesfake "k8s.io/client-go/kubernetes/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestEnsureConfigMapsRequireExistingOwner(t *testing.T) {
	ctx := context.Background()
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	owner := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "owner", UID: types.UID("owner-uid")}}
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: "aliases", Namespace: "test"},
		Data:       map[string]string{"aliases.conf": "foreign"},
	}
	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(cm).Build()
	h, err := helper.NewHelper(owner, client, kubernetesfake.NewSimpleClientset(), scheme, logr.Discard())
	if err != nil {
		t.Fatal(err)
	}
	template := util.Template{
		Name:                 cm.Name,
		Namespace:            cm.Namespace,
		Type:                 util.TemplateTypeNone,
		CustomData:           map[string]string{"aliases.conf": "managed"},
		RequireExistingOwner: true,
	}
	if err := EnsureConfigMaps(ctx, h, owner, []util.Template{template}, nil); !errors.Is(err, ErrExistingConfigMapNotOwned) {
		t.Fatalf("expected ownership conflict, got %v", err)
	}
	actual := &corev1.ConfigMap{}
	if err := client.Get(ctx, types.NamespacedName{Name: cm.Name, Namespace: cm.Namespace}, actual); err != nil {
		t.Fatal(err)
	}
	if actual.Data["aliases.conf"] != "foreign" || len(actual.OwnerReferences) != 0 {
		t.Fatalf("strict ownership changed foreign ConfigMap: %+v", actual)
	}
	invalid := template
	invalid.SkipSetOwner = true
	if err := EnsureConfigMaps(ctx, h, owner, []util.Template{invalid}, nil); err == nil {
		t.Fatal("expected strict ownership with SkipSetOwner to be rejected")
	}
	invalid = template
	invalid.Type = util.TemplateTypeCustom
	if err := EnsureConfigMaps(ctx, h, owner, []util.Template{invalid}, nil); err == nil {
		t.Fatal("expected strict ownership with a custom template to be rejected")
	}

	template.RequireExistingOwner = false
	if err := EnsureConfigMaps(ctx, h, owner, []util.Template{template}, nil); err != nil {
		t.Fatalf("default ownership behavior changed: %v", err)
	}
	if err := client.Get(ctx, types.NamespacedName{Name: cm.Name, Namespace: cm.Namespace}, actual); err != nil {
		t.Fatal(err)
	}
	if actual.Data["aliases.conf"] != "managed" || !metav1.IsControlledBy(actual, owner) {
		t.Fatalf("expected default adoption, got %+v", actual)
	}

	template.RequireExistingOwner = true
	template.CustomData["aliases.conf"] = "updated"
	if err := EnsureConfigMaps(ctx, h, owner, []util.Template{template}, nil); err != nil {
		t.Fatalf("strict ownership rejected its own ConfigMap: %v", err)
	}
	if err := client.Get(ctx, types.NamespacedName{Name: cm.Name, Namespace: cm.Namespace}, actual); err != nil {
		t.Fatal(err)
	}
	if actual.Data["aliases.conf"] != "updated" || !metav1.IsControlledBy(actual, owner) {
		t.Fatalf("strict ownership did not update its own ConfigMap: %+v", actual)
	}
}
