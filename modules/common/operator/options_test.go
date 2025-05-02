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

package operator

import (
	"strconv"
	"testing"
	"time"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
)

func TestGetEnvInDuration(t *testing.T) {
	tests := []struct {
		name     string
		envName  string
		envValue string
		wantErr  bool
		expect   time.Duration
	}{
		{
			name:     "Test valid",
			envName:  "VALID",
			envValue: "30",
			wantErr:  false,
			expect:   time.Duration(30 * time.Second),
		},
		{
			name:     "Test invvalid",
			envName:  "INVALID",
			envValue: "3x0",
			wantErr:  true,
			expect:   time.Duration(60 * time.Second),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(tt.envName, tt.envValue)
			res, err := getEnvInDuration(tt.envName)
			if tt.wantErr {
				if err == nil {
					t.Errorf("getEnvInDuration() expected error but got none")
				}
				return
			} else if err != nil {
				t.Errorf("getEnvInDuration() unexpected error: %v", err)
				return
			}
			if res != tt.expect {
				t.Errorf("getEnvInDuration() got = %v, want %v", res, tt.expect)
			}
		})
	}
}

func TestSetManagerOptions(t *testing.T) {
	var durationInt int64
	var expectedValue time.Duration
	setupLog := logr.New(nil)

	tests := []struct {
		name          string
		leaseDuration string
		renewDeadline string
		retryPeriod   string
		wantErr       bool
	}{
		{
			name:          "Test SetOperatorOptions valid values",
			leaseDuration: "137",
			renewDeadline: "107",
			retryPeriod:   "26",
			wantErr:       false,
		},
		{
			name:          "Test SetOperatorOptions invalid values",
			leaseDuration: "foo",
			renewDeadline: "bar",
			retryPeriod:   "INVALID",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("LEASE_DURATION", tt.leaseDuration)
			t.Setenv("RENEW_DEADLINE", tt.renewDeadline)
			t.Setenv("RETRY_PERIOD", tt.retryPeriod)
			options := ctrl.Options{}
			err := SetManagerOptions(&options, setupLog)
			if tt.wantErr {
				if err == nil {
					t.Errorf("SetOperatorOptions() expected error but got none")
				}
				return
			} else if err != nil {
				t.Errorf("SetOperatorOptions() unexpected error: %v", err)
				return
			}

			durationInt, _ = strconv.ParseInt(tt.leaseDuration, 10, 64)
			expectedValue = time.Duration(durationInt) * time.Second
			if *options.LeaseDuration != expectedValue {
				t.Errorf("SetOperatorOptions() got = %v, want %v", options.LeaseDuration, expectedValue)
			}

			durationInt, _ = strconv.ParseInt(tt.renewDeadline, 10, 64)
			expectedValue = time.Duration(durationInt) * time.Second
			if *options.RenewDeadline != expectedValue {
				t.Errorf("SetOperatorOptions() got = %v, want %v", options.RenewDeadline, expectedValue)
			}

			durationInt, _ = strconv.ParseInt(tt.retryPeriod, 10, 64)
			expectedValue = time.Duration(durationInt) * time.Second
			if *options.RetryPeriod != expectedValue {
				t.Errorf("SetOperatorOptions() got = %v, want %v", options.RetryPeriod, expectedValue)
			}
		})
	}
}
