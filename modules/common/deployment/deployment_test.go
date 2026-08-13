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

package deployment

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestIsReady(t *testing.T) {
	tests := []struct {
		name string
		depl appsv1.Deployment
		want bool
	}{
		{
			name: "all ready",
			depl: appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Replicas: ptr.To[int32](3),
				},
				Status: appsv1.DeploymentStatus{
					Replicas:           3,
					ReadyReplicas:      3,
					UpdatedReplicas:    3,
					ObservedGeneration: 1,
				},
			},
			want: true,
		},
		{
			name: "replicas nil",
			depl: appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{},
				Status: appsv1.DeploymentStatus{
					Replicas:           0,
					ReadyReplicas:      0,
					UpdatedReplicas:    0,
					ObservedGeneration: 0,
				},
			},
			want: false,
		},
		{
			name: "ready replicas mismatch",
			depl: appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Replicas: ptr.To[int32](3),
				},
				Status: appsv1.DeploymentStatus{
					Replicas:           3,
					ReadyReplicas:      2,
					UpdatedReplicas:    3,
					ObservedGeneration: 1,
				},
			},
			want: false,
		},
		{
			name: "updated replicas mismatch",
			depl: appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Replicas: ptr.To[int32](3),
				},
				Status: appsv1.DeploymentStatus{
					Replicas:           3,
					ReadyReplicas:      3,
					UpdatedReplicas:    2,
					ObservedGeneration: 1,
				},
			},
			want: false,
		},
		{
			name: "status replicas differs from ready replicas during rolling update",
			depl: appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Replicas: ptr.To[int32](3),
				},
				Status: appsv1.DeploymentStatus{
					Replicas:           4,
					ReadyReplicas:      3,
					UpdatedReplicas:    3,
					ObservedGeneration: 1,
				},
			},
			want: false,
		},
		{
			name: "generation mismatch",
			depl: appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Replicas: ptr.To[int32](3),
				},
				Status: appsv1.DeploymentStatus{
					Replicas:           3,
					ReadyReplicas:      3,
					UpdatedReplicas:    3,
					ObservedGeneration: 0,
				},
			},
			want: false,
		},
		{
			name: "zero replicas all match",
			depl: appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Replicas: ptr.To[int32](0),
				},
				Status: appsv1.DeploymentStatus{
					Replicas:           0,
					ReadyReplicas:      0,
					UpdatedReplicas:    0,
					ObservedGeneration: 0,
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			tt.depl.Generation = tt.depl.Status.ObservedGeneration
			if tt.name == "generation mismatch" {
				tt.depl.Generation = tt.depl.Status.ObservedGeneration + 1
			}
			g.Expect(IsReady(tt.depl)).To(Equal(tt.want))
		})
	}
}

func TestIsReadyForInput(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = appsv1.AddToScheme(scheme)

	readyDepl := func(hash string) *appsv1.Deployment {
		return &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "test-depl",
				Namespace:  "test-ns",
				Generation: 1,
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: ptr.To[int32](1),
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name: "service",
								Env: []corev1.EnvVar{
									{Name: ConfigHashEnvVar, Value: hash},
								},
							},
						},
					},
				},
			},
			Status: appsv1.DeploymentStatus{
				Replicas:           1,
				ReadyReplicas:      1,
				UpdatedReplicas:    1,
				ObservedGeneration: 1,
			},
		}
	}

	name := types.NamespacedName{Name: "test-depl", Namespace: "test-ns"}

	tests := []struct {
		name       string
		depl       *appsv1.Deployment
		configHash string
		want       bool
		wantErr    bool
	}{
		{
			name:       "ready with matching config hash",
			depl:       readyDepl("hash-abc"),
			configHash: "hash-abc",
			want:       true,
		},
		{
			name:       "ready but config hash mismatch",
			depl:       readyDepl("hash-abc"),
			configHash: "hash-xyz",
			want:       false,
		},
		{
			name: "ready but no CONFIG_HASH env var",
			depl: func() *appsv1.Deployment {
				d := readyDepl("")
				d.Spec.Template.Spec.Containers[0].Env = nil
				return d
			}(),
			configHash: "hash-abc",
			want:       false,
		},
		{
			name: "not ready - replicas mismatch",
			depl: func() *appsv1.Deployment {
				d := readyDepl("hash-abc")
				d.Status.ReadyReplicas = 0
				return d
			}(),
			configHash: "hash-abc",
			want:       false,
		},
		{
			name: "not ready - generation mismatch",
			depl: func() *appsv1.Deployment {
				d := readyDepl("hash-abc")
				d.Generation = 2
				return d
			}(),
			configHash: "hash-abc",
			want:       false,
		},
		{
			name: "not ready - rolling update in progress",
			depl: func() *appsv1.Deployment {
				d := readyDepl("hash-abc")
				d.Status.Replicas = 2
				return d
			}(),
			configHash: "hash-abc",
			want:       false,
		},
		{
			name:       "not found",
			depl:       nil,
			configHash: "hash-abc",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			builder := fake.NewClientBuilder().WithScheme(scheme)
			if tt.depl != nil {
				builder = builder.WithObjects(tt.depl)
			}
			reader := builder.Build()

			got, err := IsReadyForInput(context.Background(), reader, name, tt.configHash)
			if tt.wantErr {
				g.Expect(err).To(HaveOccurred())
			} else {
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(got).To(Equal(tt.want))
			}
		})
	}
}
