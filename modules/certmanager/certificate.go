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
	certmgrmetav1 "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	"github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	"github.com/openstack-k8s-operators/lib-common/modules/common/secret"
	"github.com/openstack-k8s-operators/lib-common/modules/common/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	k8s_corev1 "k8s.io/api/core/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
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
//	dnsNames:
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

// EnsureCert - creates a certificate for hostnames, ensures the sercret has the required key/cert and return the secret
func EnsureCert(
	ctx context.Context,
	helper *helper.Helper,
	issuerName string,
	certName string,
	duration *time.Duration,
	hostnames []string,
	annotations map[string]string,
	labels map[string]string,
) (*k8s_corev1.Secret, ctrl.Result, error) {
	// get issuer
	issuer := &certmgrv1.Issuer{}
	namespace := helper.GetBeforeObject().GetNamespace()

	err := helper.GetClient().Get(ctx, types.NamespacedName{Name: issuerName, Namespace: namespace}, issuer)
	if err != nil {
		err = fmt.Errorf("Error getting issuer %s/%s - %w", issuerName, namespace, err)

		return nil, ctrl.Result{}, err
	}

	// default the cert duration to one year (default is 90days)
	if duration == nil {
		duration = ptr.To(time.Hour * 24 * 365)
	}

	certSecretName := "cert-" + certName
	certReq := Cert(
		certName,
		namespace,
		labels,
		certmgrv1.CertificateSpec{
			DNSNames: hostnames,
			Duration: &metav1.Duration{
				Duration: *duration,
			},
			IssuerRef: certmgrmetav1.ObjectReference{
				Name:  issuer.Name,
				Kind:  issuer.Kind,
				Group: issuer.GroupVersionKind().Group,
			},
			SecretName: certSecretName,
			SecretTemplate: &certmgrv1.CertificateSecretTemplate{
				Annotations: annotations,
				Labels:      labels,
			},
			// TODO Usages, e.g. for client cert
		},
	)

	cert := NewCertificate(certReq, 5)
	ctrlResult, err := cert.CreateOrPatch(ctx, helper)
	if err != nil {
		return nil, ctrlResult, err
	} else if (ctrlResult != ctrl.Result{}) {
		return nil, ctrlResult, nil
	}

	// get cert secret
	certSecret, _, err := secret.GetSecret(ctx, helper, certSecretName, namespace)
	if err != nil {
		return nil, ctrl.Result{}, err
	}

	// check if secret has the right keys
	_, hasTLSKey := certSecret.Data["tls.key"]
	_, hasTLSCert := certSecret.Data["tls.crt"]
	if !hasTLSCert || !hasTLSKey {
		err := fmt.Errorf("TLS secret %s in namespace %s does not have the fields tls.crt and tls.key", certSecretName, namespace)
		return nil, ctrl.Result{}, err
	}

	return certSecret, ctrl.Result{}, nil
}
