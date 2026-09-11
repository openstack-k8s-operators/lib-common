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

package certmanager

import (
	"testing"

	certmgrv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	. "github.com/onsi/gomega"
)

// TestDefaultUsages verifies that the default certificate key usages are
// algorithm-aware: RSA and the empty (cert-manager default) algorithm include
// keyEncipherment, while ECDSA and Ed25519 omit it.
func TestDefaultUsages(t *testing.T) {
	rsaUsages := []certmgrv1.KeyUsage{
		certmgrv1.UsageKeyEncipherment,
		certmgrv1.UsageDigitalSignature,
		certmgrv1.UsageServerAuth,
	}
	nonRSAUsages := []certmgrv1.KeyUsage{
		certmgrv1.UsageDigitalSignature,
		certmgrv1.UsageServerAuth,
	}

	tests := []struct {
		name      string
		algorithm certmgrv1.PrivateKeyAlgorithm
		want      []certmgrv1.KeyUsage
	}{
		{
			name:      "empty algorithm defaults to RSA and includes keyEncipherment",
			algorithm: "",
			want:      rsaUsages,
		},
		{
			name:      "RSA includes keyEncipherment",
			algorithm: certmgrv1.RSAKeyAlgorithm,
			want:      rsaUsages,
		},
		{
			name:      "ECDSA omits keyEncipherment",
			algorithm: certmgrv1.ECDSAKeyAlgorithm,
			want:      nonRSAUsages,
		},
		{
			name:      "Ed25519 omits keyEncipherment",
			algorithm: certmgrv1.Ed25519KeyAlgorithm,
			want:      nonRSAUsages,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(defaultUsages(tt.algorithm)).To(Equal(tt.want))
		})
	}
}
