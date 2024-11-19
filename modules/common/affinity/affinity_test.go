/*
Copyright 2020 Red Hat

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

package affinity

import (
	"testing"

	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var affinityObj = &corev1.Affinity{
	PodAntiAffinity: &corev1.PodAntiAffinity{
		PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
			{
				PodAffinityTerm: corev1.PodAffinityTerm{
					LabelSelector: &metav1.LabelSelector{
						MatchExpressions: []metav1.LabelSelectorRequirement{
							{
								Key:      "ThisSelector",
								Operator: metav1.LabelSelectorOpIn,
								Values:   []string{"selectorValue1", "selectorValue2"},
							},
						},
					},
					TopologyKey: "ThisTopologyKey",
				},
				Weight: 100,
			},
		},
	},
}

func TestDistributePods(t *testing.T) {

	t.Run("Default pod distribution", func(t *testing.T) {
		g := NewWithT(t)

		d, _ := DistributePods("ThisSelector", []string{"selectorValue1", "selectorValue2"}, "ThisTopologyKey", nil)

		g.Expect(d).To(BeEquivalentTo(affinityObj))
	})
}
