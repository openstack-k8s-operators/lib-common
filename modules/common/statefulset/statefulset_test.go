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

package statefulset

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
		sts  appsv1.StatefulSet
		want bool
	}{
		{
			name: "all ready and revisions match",
			sts: appsv1.StatefulSet{
				Spec: appsv1.StatefulSetSpec{
					Replicas: ptr.To[int32](3),
				},
				Status: appsv1.StatefulSetStatus{
					ReadyReplicas:      3,
					UpdatedReplicas:    3,
					ObservedGeneration: 1,
					CurrentRevision:    "rev-abc",
					UpdateRevision:     "rev-abc",
				},
			},
			want: true,
		},
		{
			name: "replicas nil",
			sts: appsv1.StatefulSet{
				Spec: appsv1.StatefulSetSpec{},
				Status: appsv1.StatefulSetStatus{
					ReadyReplicas:      0,
					UpdatedReplicas:    0,
					ObservedGeneration: 0,
					CurrentRevision:    "rev-abc",
					UpdateRevision:     "rev-abc",
				},
			},
			want: false,
		},
		{
			name: "ready replicas mismatch",
			sts: appsv1.StatefulSet{
				Spec: appsv1.StatefulSetSpec{
					Replicas: ptr.To[int32](3),
				},
				Status: appsv1.StatefulSetStatus{
					ReadyReplicas:      2,
					UpdatedReplicas:    3,
					ObservedGeneration: 1,
					CurrentRevision:    "rev-abc",
					UpdateRevision:     "rev-abc",
				},
			},
			want: false,
		},
		{
			name: "updated replicas mismatch",
			sts: appsv1.StatefulSet{
				Spec: appsv1.StatefulSetSpec{
					Replicas: ptr.To[int32](3),
				},
				Status: appsv1.StatefulSetStatus{
					ReadyReplicas:      3,
					UpdatedReplicas:    2,
					ObservedGeneration: 1,
					CurrentRevision:    "rev-abc",
					UpdateRevision:     "rev-abc",
				},
			},
			want: false,
		},
		{
			name: "generation mismatch",
			sts: appsv1.StatefulSet{
				Spec: appsv1.StatefulSetSpec{
					Replicas: ptr.To[int32](3),
				},
				Status: appsv1.StatefulSetStatus{
					ReadyReplicas:      3,
					UpdatedReplicas:    3,
					ObservedGeneration: 0,
					CurrentRevision:    "rev-abc",
					UpdateRevision:     "rev-abc",
				},
			},
			want: false,
		},
		{
			name: "current revision differs from update revision",
			sts: appsv1.StatefulSet{
				Spec: appsv1.StatefulSetSpec{
					Replicas: ptr.To[int32](3),
				},
				Status: appsv1.StatefulSetStatus{
					ReadyReplicas:      3,
					UpdatedReplicas:    3,
					ObservedGeneration: 1,
					CurrentRevision:    "rev-old",
					UpdateRevision:     "rev-new",
				},
			},
			want: false,
		},
		{
			name: "zero replicas all match",
			sts: appsv1.StatefulSet{
				Spec: appsv1.StatefulSetSpec{
					Replicas: ptr.To[int32](0),
				},
				Status: appsv1.StatefulSetStatus{
					ReadyReplicas:      0,
					UpdatedReplicas:    0,
					ObservedGeneration: 0,
					CurrentRevision:    "",
					UpdateRevision:     "",
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			tt.sts.Generation = tt.sts.Status.ObservedGeneration
			if tt.name == "generation mismatch" {
				tt.sts.Generation = tt.sts.Status.ObservedGeneration + 1
			}
			g.Expect(IsReady(tt.sts)).To(Equal(tt.want))
		})
	}
}

func TestIsReadyForInput(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = appsv1.AddToScheme(scheme)

	readySTS := func(hash string) *appsv1.StatefulSet {
		sts := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "test-sts",
				Namespace:  "test-ns",
				Generation: 1,
			},
			Spec: appsv1.StatefulSetSpec{
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
			Status: appsv1.StatefulSetStatus{
				ReadyReplicas:      1,
				UpdatedReplicas:    1,
				ObservedGeneration: 1,
				CurrentRevision:    "rev-abc",
				UpdateRevision:     "rev-abc",
			},
		}
		return sts
	}

	name := types.NamespacedName{Name: "test-sts", Namespace: "test-ns"}

	tests := []struct {
		name       string
		sts        *appsv1.StatefulSet
		configHash string
		want       bool
		wantErr    bool
	}{
		{
			name:       "ready with matching config hash",
			sts:        readySTS("hash-abc"),
			configHash: "hash-abc",
			want:       true,
		},
		{
			name:       "ready but config hash mismatch",
			sts:        readySTS("hash-abc"),
			configHash: "hash-xyz",
			want:       false,
		},
		{
			name: "ready but no CONFIG_HASH env var",
			sts: func() *appsv1.StatefulSet {
				sts := readySTS("")
				sts.Spec.Template.Spec.Containers[0].Env = nil
				return sts
			}(),
			configHash: "hash-abc",
			want:       false,
		},
		{
			name: "not ready - replicas mismatch",
			sts: func() *appsv1.StatefulSet {
				sts := readySTS("hash-abc")
				sts.Status.ReadyReplicas = 0
				return sts
			}(),
			configHash: "hash-abc",
			want:       false,
		},
		{
			name: "not ready - generation mismatch",
			sts: func() *appsv1.StatefulSet {
				sts := readySTS("hash-abc")
				sts.Generation = 2
				return sts
			}(),
			configHash: "hash-abc",
			want:       false,
		},
		{
			name: "not ready - revision mismatch",
			sts: func() *appsv1.StatefulSet {
				sts := readySTS("hash-abc")
				sts.Status.CurrentRevision = "rev-old"
				return sts
			}(),
			configHash: "hash-abc",
			want:       false,
		},
		{
			name:       "not found",
			sts:        nil,
			configHash: "hash-abc",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			builder := fake.NewClientBuilder().WithScheme(scheme)
			if tt.sts != nil {
				builder = builder.WithObjects(tt.sts)
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
