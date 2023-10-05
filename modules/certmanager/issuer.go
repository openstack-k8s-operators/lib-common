/*
Copyright 2023 Red Hat

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

package certmanager

import (
	"context"
	"fmt"
	"time"

	certmgrv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	"github.com/openstack-k8s-operators/lib-common/modules/common/util"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	// RootCAIssuerInternalLabel for internal RootCA to issue internal TLS Certs
	RootCAIssuerInternalLabel = "osp-rootca-issuer-internal"
)

// Issuer -
type Issuer struct {
	issuer  *certmgrv1.Issuer
	timeout time.Duration
}

// NewIssuer returns an initialized Issuer.
func NewIssuer(
	issuer *certmgrv1.Issuer,
	timeout time.Duration,
) *Issuer {
	return &Issuer{
		issuer:  issuer,
		timeout: timeout,
	}
}

// SelfSignedIssuer returns a self signed issuer.
func SelfSignedIssuer(
	name string,
	namespace string,
	labels map[string]string,
) *certmgrv1.Issuer {
	return &certmgrv1.Issuer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: certmgrv1.IssuerSpec{
			IssuerConfig: certmgrv1.IssuerConfig{
				SelfSigned: &certmgrv1.SelfSignedIssuer{},
			},
		},
	}
}

// CAIssuer returns an CA issuer.
func CAIssuer(
	name string,
	namespace string,
	labels map[string]string,
	secretName string,
) *certmgrv1.Issuer {
	return &certmgrv1.Issuer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: certmgrv1.IssuerSpec{
			IssuerConfig: certmgrv1.IssuerConfig{
				CA: &certmgrv1.CAIssuer{
					SecretName: secretName,
				},
			},
		},
	}
}

// CreateOrPatch - creates or patches a issuer, reconciles after Xs if object won't exist.
func (i *Issuer) CreateOrPatch(
	ctx context.Context,
	h *helper.Helper,
) (ctrl.Result, error) {
	issuer := &certmgrv1.Issuer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      i.issuer.Name,
			Namespace: i.issuer.Namespace,
		},
	}

	op, err := controllerutil.CreateOrPatch(ctx, h.GetClient(), issuer, func() error {
		issuer.Labels = util.MergeStringMaps(issuer.Labels, i.issuer.Labels)
		issuer.Annotations = util.MergeStringMaps(issuer.Annotations, i.issuer.Annotations)
		issuer.Spec = i.issuer.Spec

		err := controllerutil.SetControllerReference(h.GetBeforeObject(), issuer, h.GetScheme())
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			h.GetLogger().Info(fmt.Sprintf("Issuer %s not found, reconcile in %s", issuer.Name, i.timeout))
			return ctrl.Result{RequeueAfter: i.timeout}, nil
		}
		return ctrl.Result{}, err
	}
	if op != controllerutil.OperationResultNone {
		h.GetLogger().Info(fmt.Sprintf("Route %s - %s", issuer.Name, op))
	}

	return ctrl.Result{}, nil
}

// Delete - delete an issuer.
func (i *Issuer) Delete(
	ctx context.Context,
	h *helper.Helper,
) error {

	err := h.GetClient().Delete(ctx, i.issuer)
	if err != nil && !k8s_errors.IsNotFound(err) {
		return fmt.Errorf("Error deleting issuer %s: %w", i.issuer.Name, err)
	}

	return nil
}
