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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Certificate -
type Certificate struct {
	certificate *certmgrv1.Certificate
	timeout     time.Duration
}

// NewCertificate returns an initialized Certificate.
func NewCertificate(
	certificate *certmgrv1.Certificate,
	timeout time.Duration,
) *Certificate {
	return &Certificate{
		certificate: certificate,
		timeout:     timeout,
	}
}

// Cert returns an initialized certificate request obj.
// minimal spec should be:
// Spec:
//
//	commonName: keystone-public-openstack.apps-crc.testing
//	dnsNames:
//	- keystone-public-openstack
//	- keystone-public-openstack.apps-crc.testing
//	issuerRef:
//	   kind: Issuer
//	   name: osp-rootca-issuer
//	secretName: keystone-public-cert
func Cert(
	name string,
	namespace string,
	labels map[string]string,
	spec certmgrv1.CertificateSpec,

) *certmgrv1.Certificate {
	return &certmgrv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: spec,
	}
}

// CreateOrPatch - creates or patches a certificate, reconciles after Xs if object won't exist.
func (c *Certificate) CreateOrPatch(
	ctx context.Context,
	h *helper.Helper,
) (ctrl.Result, error) {
	cert := &certmgrv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.certificate.Name,
			Namespace: c.certificate.Namespace,
		},
	}

	op, err := controllerutil.CreateOrPatch(ctx, h.GetClient(), cert, func() error {
		cert.Labels = util.MergeStringMaps(cert.Labels, c.certificate.Labels)
		cert.Annotations = util.MergeStringMaps(cert.Annotations, c.certificate.Annotations)
		cert.Spec = c.certificate.Spec

		err := controllerutil.SetControllerReference(h.GetBeforeObject(), cert, h.GetScheme())
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			h.GetLogger().Info(fmt.Sprintf("Certificate %s not found, reconcile in %s", cert.Name, c.timeout))
			return ctrl.Result{RequeueAfter: c.timeout}, nil
		}
		return ctrl.Result{}, err
	}
	if op != controllerutil.OperationResultNone {
		h.GetLogger().Info(fmt.Sprintf("Route %s - %s", cert.Name, op))
	}

	return ctrl.Result{}, nil
}

// Delete - delete a certificate.
func (c *Certificate) Delete(
	ctx context.Context,
	h *helper.Helper,
) error {

	err := h.GetClient().Delete(ctx, c.certificate)
	if err != nil && !k8s_errors.IsNotFound(err) {
		return fmt.Errorf("Error deleting certificate %s: %w", c.certificate.Name, err)
	}

	return nil
}
