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

package pdb

import (
	"context"
	"fmt"
	"time"

	"github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	"github.com/openstack-k8s-operators/lib-common/modules/common/util"
	policyv1 "k8s.io/api/policy/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// NewPDB returns an initialized PDB.
func NewPDB(
	pdb *policyv1.PodDisruptionBudget,
	timeout time.Duration,
) *PDB {
	return &PDB{
		pdb:     pdb,
		timeout: timeout,
	}
}

// MaxUnavailablePodDisruptionBudget returns a PodDisruptionBudget with the specified maxUnavailable and label selector
func MaxUnavailablePodDisruptionBudget(name, namespace string, maxUnavailable intstr.IntOrString, labelSelector map[string]string) *policyv1.PodDisruptionBudget {
	return &policyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: policyv1.PodDisruptionBudgetSpec{
			MaxUnavailable: &maxUnavailable,
			Selector: &metav1.LabelSelector{
				MatchLabels: labelSelector,
			},
		},
	}
}

// MinAvailablePodDisruptionBudget returns a PodDisruptionBudget with the specified minAvailable and label selector
func MinAvailablePodDisruptionBudget(name, namespace string, minAvailable intstr.IntOrString, labelSelector map[string]string) *policyv1.PodDisruptionBudget {
	return &policyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: policyv1.PodDisruptionBudgetSpec{
			MinAvailable: &minAvailable,
			Selector: &metav1.LabelSelector{
				MatchLabels: labelSelector,
			},
		},
	}
}

// CreateOrPatch - creates or patches a PodDisruptionBudget, reconciles after Xs if object won't exist.
func (p *PDB) CreateOrPatch(
	ctx context.Context,
	h *helper.Helper,
) (ctrl.Result, error) {
	pdb := &policyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      p.pdb.Name,
			Namespace: p.pdb.Namespace,
		},
	}

	op, err := controllerutil.CreateOrPatch(ctx, h.GetClient(), pdb, func() error {
		pdb.Labels = util.MergeStringMaps(pdb.Labels, p.pdb.Labels)
		pdb.Annotations = util.MergeStringMaps(pdb.Annotations, p.pdb.Annotations)
		pdb.Spec = p.pdb.Spec

		err := controllerutil.SetControllerReference(h.GetBeforeObject(), pdb, h.GetScheme())
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			h.GetLogger().Info(fmt.Sprintf("PodDisruptionBudget %s not found, reconcile in %s", pdb.Name, p.timeout))
			return ctrl.Result{RequeueAfter: p.timeout}, nil
		}
		return ctrl.Result{}, err
	}
	if op != controllerutil.OperationResultNone {
		h.GetLogger().Info(fmt.Sprintf("PodDisruptionBudget %s - %s", pdb.Name, op))
	}

	// update the pdb object of the pdb type
	p.pdb, err = GetPDBWithName(ctx, h, pdb.GetName(), pdb.GetNamespace())
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// Delete - delete a PodDisruptionBudget.
func (p *PDB) Delete(
	ctx context.Context,
	h *helper.Helper,
) error {
	err := h.GetClient().Delete(ctx, p.pdb)
	if err != nil && !k8s_errors.IsNotFound(err) {
		return fmt.Errorf("error deleting PodDisruptionBudget %s: %w", p.pdb.Name, err)
	}

	return nil
}

// GetPDB - get the PodDisruptionBudget object.
func (p *PDB) GetPDB() policyv1.PodDisruptionBudget {
	return *p.pdb
}

// GetPDBWithName func
func GetPDBWithName(
	ctx context.Context,
	h *helper.Helper,
	name string,
	namespace string,
) (*policyv1.PodDisruptionBudget, error) {

	pdb := &policyv1.PodDisruptionBudget{}
	err := h.GetClient().Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, pdb)
	if err != nil {
		return pdb, err
	}

	return pdb, nil
}

// DeletePDBWithName deletes a PodDisruptionBudget by name and namespace
func DeletePDBWithName(
	ctx context.Context,
	h *helper.Helper,
	name string,
	namespace string,
) error {
	pdb := &policyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	err := h.GetClient().Delete(ctx, pdb)
	if err != nil && !k8s_errors.IsNotFound(err) {
		return fmt.Errorf("error deleting PodDisruptionBudget %s/%s: %w", namespace, name, err)
	}

	return nil
}

// IsReady - validates when PodDisruptionBudget is ready
// - returns true if the PDB has been processed by the controller and has valid status
func IsReady(pdb policyv1.PodDisruptionBudget) bool {
	return pdb.Status.ObservedGeneration == pdb.Generation &&
		pdb.Status.ExpectedPods > 0 &&
		pdb.Status.CurrentHealthy >= 0
}
