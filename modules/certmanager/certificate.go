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
	"sort"
	"time"

	certmgrv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	certmgrmetav1 "github.com/cert-manager/cert-manager/pkg/apis/meta/v1"
	"github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	"github.com/openstack-k8s-operators/lib-common/modules/common/net"
	"github.com/openstack-k8s-operators/lib-common/modules/common/secret"
	"github.com/openstack-k8s-operators/lib-common/modules/common/service"
	"github.com/openstack-k8s-operators/lib-common/modules/common/util"

	"golang.org/x/exp/maps"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	k8s_corev1 "k8s.io/api/core/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
)

// Certificate -
type Certificate struct {
	certificate *certmgrv1.Certificate
	timeout     time.Duration
}

// CertificateRequest -
type CertificateRequest struct {
	IssuerName  string
	CertName    string
	CommonName  *string
	Duration    *time.Duration
	RenewBefore *time.Duration
	Hostnames   []string
	Ips         []string
	Annotations map[string]string
	Labels      map[string]string
	Usages      []certmgrv1.KeyUsage
	Subject     *certmgrv1.X509Subject
}

// NewCertificate returns an initialized Certificate.
func NewCertificate(
	certificate *certmgrv1.Certificate,
	timeout time.Duration,
) *Certificate {
	crt := &Certificate{
		certificate: certificate,
		timeout:     timeout,
	}

	crt.certificate.Spec.IPAddresses = net.SortIPs(crt.certificate.Spec.IPAddresses)
	sort.Strings(crt.certificate.Spec.DNSNames)

	return crt
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
	owner client.Object,
) (ctrl.Result, controllerutil.OperationResult, error) {
	var err error
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

		if owner != nil {
			err = controllerutil.SetControllerReference(owner, cert, h.GetScheme())
		} else {
			err = controllerutil.SetControllerReference(h.GetBeforeObject(), cert, h.GetScheme())
		}
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			h.GetLogger().Info(fmt.Sprintf("Certificate %s not found, reconcile in %s", cert.Name, c.timeout))
			return ctrl.Result{RequeueAfter: c.timeout}, op, nil
		}
		return ctrl.Result{}, op, err
	}
	if op != controllerutil.OperationResultNone {
		h.GetLogger().Info(fmt.Sprintf("Route %s - %s", cert.Name, op))
	}

	return ctrl.Result{}, op, nil
}

// Delete - delete a certificate.
func (c *Certificate) Delete(
	ctx context.Context,
	h *helper.Helper,
) error {

	err := h.GetClient().Delete(ctx, c.certificate)
	if err != nil && !k8s_errors.IsNotFound(err) {
		return fmt.Errorf("error deleting certificate %s: %w", c.certificate.Name, err)
	}

	return nil
}

// EnsureCert - creates a certificate, ensures the secret has the required key/cert and return the secret
func EnsureCert(
	ctx context.Context,
	helper *helper.Helper,
	request CertificateRequest,
	owner client.Object,
) (*k8s_corev1.Secret, ctrl.Result, error) {
	// get issuer
	issuer := &certmgrv1.Issuer{}
	namespace := helper.GetBeforeObject().GetNamespace()

	err := helper.GetClient().Get(ctx, types.NamespacedName{Name: request.IssuerName, Namespace: namespace}, issuer)
	if err != nil {
		err = fmt.Errorf("error getting issuer %s/%s - %w", request.IssuerName, namespace, err)

		return nil, ctrl.Result{}, err
	}

	// default the cert duration to one year (default is 90days)
	if request.Duration == nil {
		request.Duration = ptr.To(time.Hour * 24 * 365)
	}

	// default to serverAuth
	if request.Usages == nil {
		request.Usages = []certmgrv1.KeyUsage{
			certmgrv1.UsageKeyEncipherment,
			certmgrv1.UsageDigitalSignature,
			certmgrv1.UsageServerAuth,
		}
	}

	certSecretName := "cert-" + request.CertName
	certSpec := certmgrv1.CertificateSpec{
		Duration: &metav1.Duration{
			Duration: *request.Duration,
		},
		IssuerRef: certmgrmetav1.ObjectReference{
			Name:  issuer.Name,
			Kind:  issuer.Kind,
			Group: issuer.GroupVersionKind().Group,
		},
		SecretName: certSecretName,
		SecretTemplate: &certmgrv1.CertificateSecretTemplate{
			Annotations: request.Annotations,
			Labels:      request.Labels,
		},
		Subject: request.Subject,
		Usages:  request.Usages,
	}

	if request.RenewBefore != nil {
		certSpec.RenewBefore = &metav1.Duration{
			Duration: *request.RenewBefore,
		}
	}

	if request.Hostnames != nil {
		certSpec.DNSNames = request.Hostnames
	}

	if request.Ips != nil {
		certSpec.IPAddresses = request.Ips
	}

	if request.CommonName != nil {
		certSpec.CommonName = *request.CommonName
	}

	certReq := Cert(
		request.CertName,
		namespace,
		request.Labels,
		certSpec,
	)

	cert := NewCertificate(certReq, 5)
	ctrlResult, op, err := cert.CreateOrPatch(ctx, helper, owner)
	if err != nil {
		return nil, ctrlResult, err
	} else if (ctrlResult != ctrl.Result{}) {
		return nil, ctrlResult, nil
	}

	// get cert secret
	certSecret, _, err := secret.GetSecret(ctx, helper, certSecretName, namespace)
	if err != nil {
		if k8s_errors.IsNotFound(err) && op == controllerutil.OperationResultCreated {
			helper.GetLogger().Info(fmt.Sprintf("Secret %s not found, reconcile in %s", certSecretName, cert.timeout))
			return nil, ctrl.Result{RequeueAfter: cert.timeout}, nil
		}
		return nil, ctrl.Result{}, err
	}

	// check if secret has the right keys
	_, hasTLSKey := certSecret.Data["tls.key"]
	_, hasTLSCert := certSecret.Data["tls.crt"]
	if !hasTLSCert || !hasTLSKey {
		err := fmt.Errorf("%w: TLS secret %s in namespace %s does not have the fields tls.crt and tls.key", util.ErrFieldNotFound, certSecretName, namespace)
		return nil, ctrl.Result{}, err
	}

	return certSecret, ctrl.Result{}, nil
}

// EnsureCertForServicesWithSelector - creates certificate for k8s services identified
// by a label selector
func EnsureCertForServicesWithSelector(
	ctx context.Context,
	helper *helper.Helper,
	namespace string,
	selector map[string]string,
	issuer string,
	owner client.Object,
) (map[string]string, ctrl.Result, error) {
	certs := map[string]string{}
	svcs, err := service.GetServicesListWithLabel(
		ctx,
		helper,
		namespace,
		selector,
	)
	if err != nil {
		return certs, ctrl.Result{}, err
	}

	for _, svc := range svcs.Items {
		hostname := fmt.Sprintf("%s.%s.svc", svc.Name, namespace)
		// create cert for the service
		certRequest := CertificateRequest{
			IssuerName: issuer,
			CertName:   fmt.Sprintf("%s-svc", svc.Name),
			Hostnames:  []string{hostname},
			Labels:     svc.Labels,
		}
		certSecret, ctrlResult, err := EnsureCert(
			ctx,
			helper,
			certRequest,
			owner)
		if err != nil {
			return certs, ctrlResult, err
		} else if (ctrlResult != ctrl.Result{}) {
			return certs, ctrlResult, nil
		}

		certs[hostname] = certSecret.Name
	}

	return certs, ctrl.Result{}, nil
}

// EnsureCertForServiceWithSelector - creates certificate for a k8s service identified
// by a label selector. The label selector must match a single service
// Note: Returns an NotFound error if <1 or >1 service found using the selector
func EnsureCertForServiceWithSelector(
	ctx context.Context,
	helper *helper.Helper,
	namespace string,
	selector map[string]string,
	issuer string,
	owner client.Object,
) (string, ctrl.Result, error) {
	var cert string
	svcs, err := service.GetServicesListWithLabel(
		ctx,
		helper,
		namespace,
		selector,
	)
	if err != nil {
		return cert, ctrl.Result{}, err
	}
	if len(svcs.Items) != 1 {
		err = k8s_errors.NewNotFound(
			schema.GroupResource{Group: "", Resource: "services"},
			fmt.Sprintf("%d services identified by selector: %+v", len(svcs.Items), selector))

		return cert, ctrl.Result{}, err
	}

	certs, ctrlResult, err := EnsureCertForServicesWithSelector(
		ctx, helper, namespace, selector, issuer, owner)
	if err != nil {
		return cert, ctrlResult, err
	} else if (ctrlResult != ctrl.Result{}) {
		return cert, ctrlResult, nil
	}

	hostname := maps.Keys(certs)

	return certs[hostname[0]], ctrl.Result{}, nil
}
