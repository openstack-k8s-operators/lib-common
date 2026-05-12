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

package edpm

import (
	"context"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	oko_secret "github.com/openstack-k8s-operators/lib-common/modules/common/secret"
)

func makeNodeSet(name, namespace string, secretHashes map[string]string) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(NodeSetGVK)
	obj.SetName(name)
	obj.SetNamespace(namespace)
	if len(secretHashes) > 0 {
		hashes := map[string]interface{}{}
		for k, v := range secretHashes {
			hashes[k] = v
		}
		obj.Object["status"] = map[string]interface{}{
			"secretHashes": hashes,
		}
	}
	return obj
}

func newTestSchemeAndMapper() (*runtime.Scheme, meta.RESTMapper) {
	s := runtime.NewScheme()
	_ = corev1.AddToScheme(s)

	mapper := meta.NewDefaultRESTMapper([]schema.GroupVersion{
		{Group: "dataplane.openstack.org", Version: "v1beta1"},
	})
	mapper.Add(NodeSetGVK, meta.RESTScopeNamespace)

	return s, mapper
}

func TestAreSecretHashesInSync_CRDNotInstalled(t *testing.T) {
	s := runtime.NewScheme()
	_ = corev1.AddToScheme(s)

	// Empty mapper — NodeSet GVK is not registered, simulating a cluster
	// where the OpenStackDataPlaneNodeSet CRD is not installed.
	mapper := meta.NewDefaultRESTMapper([]schema.GroupVersion{})

	c := fake.NewClientBuilder().
		WithScheme(s).
		WithRESTMapper(mapper).
		Build()

	inSync, info, err := AreSecretHashesInSync(context.Background(), c, "test")
	if err != nil {
		t.Errorf("AreSecretHashesInSync() unexpected error: %v", err)
	}
	if !inSync {
		t.Errorf("AreSecretHashesInSync() inSync = false, want true when CRD not installed (info: %s)", info)
	}
}

func TestAreSecretHashesInSync(t *testing.T) {
	currentSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nova-cell1-compute-config",
			Namespace: "test",
		},
		Data: map[string][]byte{"transport_url": []byte("rabbit://nova:current-password@rabbitmq:5672/")},
	}
	currentHash, _ := oko_secret.Hash(currentSecret)

	tests := []struct {
		name           string
		nodesets       []*unstructured.Unstructured
		secrets        []*corev1.Secret
		wantInSync     bool
		wantInfoSubstr string
		wantErr        bool
	}{
		{
			name:       "no nodesets exist",
			wantInSync: true,
		},
		{
			name: "nodeset with stale secrets is out of sync",
			nodesets: []*unstructured.Unstructured{
				makeNodeSet("test-nodeset", "test", map[string]string{
					"nova-cell1-compute-config": "old-stale-hash",
				}),
			},
			secrets:        []*corev1.Secret{currentSecret},
			wantInSync:     false,
			wantInfoSubstr: "has changed since last full deployment",
		},
		{
			name: "nodeset with current secrets is in sync",
			nodesets: []*unstructured.Unstructured{
				makeNodeSet("test-nodeset", "test", map[string]string{
					"nova-cell1-compute-config": currentHash,
				}),
			},
			secrets:    []*corev1.Secret{currentSecret},
			wantInSync: true,
		},
		{
			name: "nodeset with empty SecretHashes is in sync",
			nodesets: []*unstructured.Unstructured{
				makeNodeSet("never-deployed-nodeset", "test", map[string]string{}),
			},
			wantInSync: true,
		},
		{
			name: "deployed secret deleted is out of sync",
			nodesets: []*unstructured.Unstructured{
				makeNodeSet("test-nodeset", "test", map[string]string{
					"deleted-secret": "some-hash",
				}),
			},
			secrets:        []*corev1.Secret{},
			wantInSync:     false,
			wantInfoSubstr: "no longer exists",
		},
		{
			name: "transport URL secret updated after credential change is out of sync",
			nodesets: []*unstructured.Unstructured{
				makeNodeSet("edpm-compute", "test", map[string]string{
					"nova-cell1-transport": "hash-with-old-credentials",
				}),
			},
			secrets: []*corev1.Secret{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "nova-cell1-transport", Namespace: "test"},
					Data:       map[string][]byte{"transport_url": []byte("rabbit://novacell2:newpass@rabbitmq:5672/")},
				},
			},
			wantInSync:     false,
			wantInfoSubstr: "has changed since last full deployment",
		},
		{
			name: "multiple nodesets - one stale blocks sync",
			nodesets: []*unstructured.Unstructured{
				makeNodeSet("up-to-date-nodeset", "test", map[string]string{
					"nova-cell1-compute-config": currentHash,
				}),
				makeNodeSet("stale-nodeset", "test", map[string]string{
					"nova-cell1-compute-config": "old-hash-from-previous-deployment",
				}),
			},
			secrets:        []*corev1.Secret{currentSecret},
			wantInSync:     false,
			wantInfoSubstr: "has changed since last full deployment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, mapper := newTestSchemeAndMapper()

			builder := fake.NewClientBuilder().
				WithScheme(s).
				WithRESTMapper(mapper)

			for _, ns := range tt.nodesets {
				builder = builder.WithObjects(ns)
			}
			for _, sec := range tt.secrets {
				builder = builder.WithObjects(sec)
			}

			c := builder.Build()

			inSync, info, err := AreSecretHashesInSync(
				context.Background(),
				c,
				"test",
			)

			if (err != nil) != tt.wantErr {
				t.Errorf("AreSecretHashesInSync() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if inSync != tt.wantInSync {
				t.Errorf("AreSecretHashesInSync() inSync = %v, want %v (info: %s)", inSync, tt.wantInSync, info)
			}

			if tt.wantInfoSubstr != "" {
				if info == "" {
					t.Errorf("AreSecretHashesInSync() info is empty, want substring %q", tt.wantInfoSubstr)
				} else if !strings.Contains(info, tt.wantInfoSubstr) {
					t.Errorf("AreSecretHashesInSync() info = %q, want substring %q", info, tt.wantInfoSubstr)
				}
			}
		})
	}
}
