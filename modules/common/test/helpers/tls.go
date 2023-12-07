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

package helpers

import (
	"k8s.io/apimachinery/pkg/types"

	corev1 "k8s.io/api/core/v1"
)

// CreateCABundleSecret creates a new secret holding the tls-ca-bundle.pem with fake data.
//
// Example usage:
//
//	caBundleSecretName = types.NamespacedName{Name: "bundlename", Namespace: namespace}
//	s := th.CreateCABundleSecret(caBundleSecretName)
func (tc *TestHelper) CreateCABundleSecret(name types.NamespacedName) *corev1.Secret {
	data := map[string][]byte{
		"tls-ca-bundle.pem": []byte("Zm9v"),
	}

	return tc.CreateSecret(name, data)
}

// CreateCertSecret creates a new secret with entries for the cert, key and a ca using fake data.
//
// Example usage:
//
//	certSecretName = types.NamespacedName{Name: "secretname", Namespace: namespace}
//	s := th.CreateCertSecret(certSecretName)
func (tc *TestHelper) CreateCertSecret(name types.NamespacedName) *corev1.Secret {
	data := map[string][]byte{
		"ca.crt":  []byte("Zm9v"),
		"tls.crt": []byte("Zm9v"),
		"tls.key": []byte("Zm9v"),
	}

	return tc.CreateSecret(name, data)
}
