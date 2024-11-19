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
	"encoding/json"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
)

// DistributePods - returns rule to ensure that two replicas of the same selector
// should not run if possible on the same worker node
func DistributePods(
	selectorKey string,
	selectorValues []string,
	topologyKey string,
	overrides *Overrides,
) (*corev1.Affinity, error) {
	// By default apply an anti-affinity policy using corev1.LabelHostname as
	// preferred scheduling policy: this maintains backward compatibility with
	// an already deployed environment
	defaultAffinity := DefaultAffinity(
		Rules{
			SelectorKey:    selectorKey,
			SelectorValues: selectorValues,
			TopologyKey:    topologyKey,
			Weight:         DefaultPreferredWeight,
		},
	)
	if overrides == nil || (overrides.Affinity == nil && overrides.AntiAffinity == nil) {
		return defaultAffinity, nil
	}

	affinityPatch := corev1.Affinity{}
	if overrides.Affinity != nil {
		affinityPatch = NewAffinity(overrides.Affinity)
	}

	antiAffinityPatch := corev1.Affinity{}
	if overrides.AntiAffinity != nil {
		antiAffinityPatch = NewAntiAffinity(overrides.AntiAffinity)
	}

	overridesSpec := &OverrideSpec{
		PodAffinity:     affinityPatch.PodAffinity,
		PodAntiAffinity: antiAffinityPatch.PodAntiAffinity,
	}

	// patch the default affinity Object with the data passed as input
	patchedAffinity, err := toCoreAffinity(defaultAffinity, overridesSpec)
	return patchedAffinity, err
}

// toCoreAffinity -
func toCoreAffinity(
	affinity *corev1.Affinity,
	override *OverrideSpec,
) (*corev1.Affinity, error) {
	aff := &corev1.Affinity{
		PodAntiAffinity: affinity.PodAntiAffinity,
		PodAffinity:     affinity.PodAffinity,
	}
	if override != nil {
		if override != nil {
			origAffinit, err := json.Marshal(affinity)
			if err != nil {
				return aff, fmt.Errorf("error marshalling Affinity Spec: %w", err)
			}
			patch, err := json.Marshal(override)
			if err != nil {
				return aff, fmt.Errorf("error marshalling Affinity Spec: %w", err)
			}
			patchedJSON, err := strategicpatch.StrategicMergePatch(origAffinit, patch, corev1.Affinity{})
			if err != nil {
				return aff, fmt.Errorf("error patching Affinity Spec: %w", err)
			}
			patchedSpec := corev1.Affinity{}
			err = json.Unmarshal(patchedJSON, &patchedSpec)
			if err != nil {
				return aff, fmt.Errorf("error unmarshalling patched Service Spec: %w", err)
			}
			aff = &patchedSpec
		}
	}
	return aff, nil
}

// WeightedPodAffinityTerm - returns a WeightedPodAffinityTerm that is assigned
// to the Affinity or AntiAffinity rule
func (affinity *Rules) WeightedPodAffinityTerm() []corev1.WeightedPodAffinityTerm {
	if affinity == nil {
		return []corev1.WeightedPodAffinityTerm{}
	}
	affinityTerm := []corev1.WeightedPodAffinityTerm{
		{
			Weight: affinity.Weight,
			PodAffinityTerm: corev1.PodAffinityTerm{
				LabelSelector: &metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{
						{
							Key:      affinity.SelectorKey,
							Operator: metav1.LabelSelectorOpIn,
							Values:   affinity.SelectorValues,
						},
					},
				},
				TopologyKey: affinity.TopologyKey,
			},
		},
	}
	return affinityTerm
}

// PodAffinityTerm -
func (affinity *Rules) PodAffinityTerm() []corev1.PodAffinityTerm {
	if affinity == nil {
		return []corev1.PodAffinityTerm{}
	}
	affinityTerm := []corev1.PodAffinityTerm{
		{
			LabelSelector: &metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{
						Key:      affinity.SelectorKey,
						Operator: metav1.LabelSelectorOpIn,
						Values:   affinity.SelectorValues,
					},
				},
			},
			TopologyKey: affinity.TopologyKey,
		},
	}
	return affinityTerm
}

// NewAffinity -
func NewAffinity(p *PodScheduling) corev1.Affinity {
	aff := &corev1.Affinity{
		PodAffinity: &corev1.PodAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution:  p.RequiredScheduling.PodAffinityTerm(),
			PreferredDuringSchedulingIgnoredDuringExecution: p.PreferredScheduling.WeightedPodAffinityTerm(),
		},
	}
	return *aff
}

// NewAntiAffinity -
func NewAntiAffinity(p *PodScheduling) corev1.Affinity {
	aff := &corev1.Affinity{
		PodAntiAffinity: &corev1.PodAntiAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution:  p.RequiredScheduling.PodAffinityTerm(),
			PreferredDuringSchedulingIgnoredDuringExecution: p.PreferredScheduling.WeightedPodAffinityTerm(),
		},
	}
	return *aff
}

// DefaultAffinity -
func DefaultAffinity(aff Rules) *corev1.Affinity {
	return &corev1.Affinity{
		PodAntiAffinity: &corev1.PodAntiAffinity{
			// This rule ensures that two replicas of the same selector
			// should not run if possible on the same worker node
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
				{
					PodAffinityTerm: corev1.PodAffinityTerm{
						LabelSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      aff.SelectorKey,
									Operator: metav1.LabelSelectorOpIn,
									Values:   aff.SelectorValues,
								},
							},
						},
						// usually corev1.LabelHostname "kubernetes.io/hostname"
						// https://github.com/kubernetes/api/blob/master/core/v1/well_known_labels.go#L20
						TopologyKey: aff.TopologyKey,
					},
					Weight: aff.Weight,
				},
			},
		},
	}
}
