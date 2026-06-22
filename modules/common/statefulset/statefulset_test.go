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
	"testing"

	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/utils/ptr"
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
