/*
Copyright 2025 Red Hat

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

package backup

import (
	"context"
	"testing"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestBuildCRDLabelCache(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = apiextensionsv1.AddToScheme(scheme)

	tests := []struct {
		name    string
		crds    []apiextensionsv1.CustomResourceDefinition
		want    CRDLabelCache
		wantErr bool
	}{
		{
			name: "CRD with backup labels",
			crds: []apiextensionsv1.CustomResourceDefinition{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "keystoneapis.keystone.openstack.org",
						Labels: map[string]string{
							BackupRestoreLabel:      "true",
							BackupRestoreOrderLabel: RestoreOrder30,
							BackupCategoryLabel:     CategoryControlPlane,
						},
					},
				},
			},
			want: CRDLabelCache{
				"keystoneapis.keystone.openstack.org": {
					Enabled:      true,
					RestoreOrder: RestoreOrder30,
					Category:     CategoryControlPlane,
				},
			},
			wantErr: false,
		},
		{
			name: "CRD without backup-restore label",
			crds: []apiextensionsv1.CustomResourceDefinition{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "other.example.com",
						Labels: map[string]string{
							"some-other": "label",
						},
					},
				},
			},
			want:    CRDLabelCache{},
			wantErr: false,
		},
		{
			name: "CRD with backup-restore=false",
			crds: []apiextensionsv1.CustomResourceDefinition{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "disabled.example.com",
						Labels: map[string]string{
							BackupRestoreLabel: "false",
						},
					},
				},
			},
			want:    CRDLabelCache{},
			wantErr: false,
		},
		{
			name: "Multiple CRDs with different configurations",
			crds: []apiextensionsv1.CustomResourceDefinition{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "keystoneapis.keystone.openstack.org",
						Labels: map[string]string{
							BackupRestoreLabel:      "true",
							BackupRestoreOrderLabel: RestoreOrder30,
							BackupCategoryLabel:     CategoryControlPlane,
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "openstackdataplaneservices.dataplane.openstack.org",
						Labels: map[string]string{
							BackupRestoreLabel:      "true",
							BackupRestoreOrderLabel: RestoreOrder60,
							BackupCategoryLabel:     CategoryDataPlane,
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "ignored.example.com",
						Labels: map[string]string{
							"other": "label",
						},
					},
				},
			},
			want: CRDLabelCache{
				"keystoneapis.keystone.openstack.org": {
					Enabled:      true,
					RestoreOrder: RestoreOrder30,
					Category:     CategoryControlPlane,
				},
				"openstackdataplaneservices.dataplane.openstack.org": {
					Enabled:      true,
					RestoreOrder: RestoreOrder60,
					Category:     CategoryDataPlane,
				},
			},
			wantErr: false,
		},
		{
			name: "CRD without category label",
			crds: []apiextensionsv1.CustomResourceDefinition{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "secrets.core",
						Labels: map[string]string{
							BackupRestoreLabel:      "true",
							BackupRestoreOrderLabel: RestoreOrder10,
						},
					},
				},
			},
			want: CRDLabelCache{
				"secrets.core": {
					Enabled:      true,
					RestoreOrder: RestoreOrder10,
					Category:     "",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			objs := make([]runtime.Object, len(tt.crds))
			for i := range tt.crds {
				objs[i] = &tt.crds[i]
			}

			c := fake.NewClientBuilder().
				WithScheme(scheme).
				WithRuntimeObjects(objs...).
				Build()

			got, err := BuildCRDLabelCache(context.Background(), c)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildCRDLabelCache() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("BuildCRDLabelCache() returned %d entries, want %d", len(got), len(tt.want))
			}

			for name, wantConfig := range tt.want {
				gotConfig, ok := got[name]
				if !ok {
					t.Errorf("BuildCRDLabelCache() missing entry for %q", name)
					continue
				}
				if gotConfig != wantConfig {
					t.Errorf("BuildCRDLabelCache()[%q] = %+v, want %+v", name, gotConfig, wantConfig)
				}
			}
		})
	}
}

func TestGetConfig(t *testing.T) {
	cache := CRDLabelCache{
		"keystoneapis.keystone.openstack.org": {
			Enabled:      true,
			RestoreOrder: RestoreOrder30,
			Category:     CategoryControlPlane,
		},
		"openstackdataplaneservices.dataplane.openstack.org": {
			Enabled:      true,
			RestoreOrder: RestoreOrder60,
			Category:     CategoryDataPlane,
		},
	}

	tests := []struct {
		name    string
		crdName string
		want    Config
	}{
		{
			name:    "existing CRD",
			crdName: "keystoneapis.keystone.openstack.org",
			want: Config{
				Enabled:      true,
				RestoreOrder: RestoreOrder30,
				Category:     CategoryControlPlane,
			},
		},
		{
			name:    "non-existent CRD",
			crdName: "unknown.example.com",
			want: Config{
				Enabled: false,
			},
		},
		{
			name:    "dataplane CRD",
			crdName: "openstackdataplaneservices.dataplane.openstack.org",
			want: Config{
				Enabled:      true,
				RestoreOrder: RestoreOrder60,
				Category:     CategoryDataPlane,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cache.GetConfig(tt.crdName)
			if got != tt.want {
				t.Errorf("GetConfig(%q) = %+v, want %+v", tt.crdName, got, tt.want)
			}
		})
	}
}
