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

package probes

import (
	"testing"

	v1 "k8s.io/api/core/v1"
)

func TestSetProbes(t *testing.T) {
	tests := []struct {
		name    string
		port    int
		disable bool
		wantURI v1.URIScheme
		wantErr bool
	}{
		{
			name:    "Disable NonTLS Listeners",
			port:    8080,
			disable: true,
			wantURI: v1.URISchemeHTTPS,
		},
		{
			name:    "Enable NonTLS Listeners",
			port:    8080,
			disable: false,
			wantURI: v1.URISchemeHTTP,
		},
		{
			name:    "Negative Port",
			port:    -8080,
			disable: false,
			wantErr: true,
		},
		{
			name:    "Port Larger than 65535",
			port:    70000,
			disable: false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			liveness, _, err := SetProbes(tt.port, tt.disable, ProbeConfig{})

			if tt.wantErr {
				if err == nil {
					t.Errorf("SetProbes() expected error but got none")
				}
				return
			} else if err != nil {
				t.Errorf("SetProbes() unexpected error: %v", err)
				return
			}
			// Only check the liveness probe if there was no error
			if liveness.HTTPGet.Scheme != tt.wantURI {
				t.Errorf("SetProbes() got = %v, want %v", liveness.HTTPGet.Scheme, tt.wantURI)
			}
		})
	}
}
