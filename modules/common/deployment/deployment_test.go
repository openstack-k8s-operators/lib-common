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
	"testing"

	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/utils/ptr"
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
