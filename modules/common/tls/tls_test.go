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

package tls

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/openstack-k8s-operators/lib-common/modules/common/deployment"
	"github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	k8sClient client.Client
)

func TestMain(m *testing.M) {
	t := &envtest.Environment{}

	cfg, err := t.Start()
	if err != nil {
		panic(err)
	}

	k8sClient, err = client.New(cfg, client.Options{})
	if err != nil {
		panic(err)
	}

	code := m.Run()

	t.Stop()

	os.Exit(code)
}

func TestCreateVolumeMounts(t *testing.T) {
	tests := []struct {
		name          string
		service       *Service
		ca            *Ca
		wantMountsLen int
	}{
		{
			name:          "No Secrets",
			service:       &Service{},
			ca:            &Ca{},
			wantMountsLen: 0,
		},
		{
			name:          "Only TLS Secret",
			service:       &Service{SecretName: "test-tls-secret"},
			ca:            &Ca{},
			wantMountsLen: 2,
		},
		{
			name:          "Only CA Secret",
			service:       &Service{},
			ca:            &Ca{CaSecretName: "test-ca1"},
			wantMountsLen: 1,
		},
		{
			name:          "TLS and CA Secrets",
			service:       &Service{SecretName: "test-tls-secret"},
			ca:            &Ca{CaSecretName: "test-ca1"},
			wantMountsLen: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tlsInstance := &TLS{Service: tt.service, Ca: tt.ca}
			mounts := tlsInstance.CreateVolumeMounts()
			if len(mounts) != tt.wantMountsLen {
				t.Errorf("CreateVolumeMounts() got = %v mounts, want %v mounts", len(mounts), tt.wantMountsLen)
			}
		})
	}
}

func TestCreateVolumes(t *testing.T) {
	tests := []struct {
		name       string
		service    *Service
		ca         *Ca
		wantVolLen int
	}{
		{
			name:       "No Secrets",
			service:    &Service{},
			ca:         &Ca{},
			wantVolLen: 0,
		},
		{
			name:       "Only TLS Secret",
			service:    &Service{SecretName: "test-tls-secret"},
			ca:         &Ca{},
			wantVolLen: 1,
		},
		{
			name:       "Only CA Secret",
			service:    &Service{},
			ca:         &Ca{CaSecretName: "test-ca1"},
			wantVolLen: 1,
		},
		{
			name:       "TLS and CA Secrets",
			service:    &Service{SecretName: "test-tls-secret"},
			ca:         &Ca{CaSecretName: "test-ca1"},
			wantVolLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tlsInstance := &TLS{Service: tt.service, Ca: tt.ca}
			volumes := tlsInstance.CreateVolumes()
			if len(volumes) != tt.wantVolLen {
				t.Errorf("CreateVolumes() got = %v volumes, want %v volumes", len(volumes), tt.wantVolLen)
			}
		})
	}
}

func TestGenerateTLSConnectionConfig(t *testing.T) {
	tests := []struct {
		name         string
		service      *Service
		ca           *Ca
		wantStmts    []string
		excludeStmts []string
	}{
		{
			name:         "No Secrets",
			service:      &Service{},
			ca:           &Ca{},
			wantStmts:    []string{},
			excludeStmts: []string{"ssl=1", "ssl-cert=", "ssl-key=", "ssl-ca="},
		},
		{
			name:         "Only TLS Secret",
			service:      &Service{SecretName: "test-tls-secret"},
			ca:           &Ca{},
			wantStmts:    []string{"ssl=1", "ssl-cert=", "ssl-key="},
			excludeStmts: []string{"ssl-ca="},
		},
		{
			name:         "Only CA Secret",
			service:      &Service{},
			ca:           &Ca{CaSecretName: "test-ca1"},
			wantStmts:    []string{"ssl=1", "ssl-ca="},
			excludeStmts: []string{"ssl-cert=", "ssl-key="},
		},
		{
			name:         "TLS and CA Secrets",
			service:      &Service{SecretName: "test-tls-secret"},
			ca:           &Ca{CaSecretName: "test-ca1"},
			wantStmts:    []string{"ssl=1", "ssl-cert=", "ssl-key=", "ssl-ca="},
			excludeStmts: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tlsInstance := &TLS{Service: tt.service, Ca: tt.ca}
			configStr := tlsInstance.CreateDatabaseClientConfig()
			var missingStmts []string
			for _, stmt := range tt.wantStmts {
				if !strings.Contains(configStr, stmt) {
					missingStmts = append(missingStmts, stmt)
				}
			}
			var unexpectedStmts []string
			for _, stmt := range tt.excludeStmts {
				if strings.Contains(configStr, stmt) {
					unexpectedStmts = append(unexpectedStmts, stmt)
				}
			}
			if len(missingStmts) != 0 || len(unexpectedStmts) != 0 {
				t.Errorf("CreateDatabaseClientConfig() "+
					"missing statements: %v, unexpected statements: %v",
					missingStmts, unexpectedStmts)
			}
		})
	}
}

func TestUpdateDeploymentWithTLS(t *testing.T) {
	assert := assert.New(t)

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment",
			Namespace: "default",
		},
	}

	customDeployment := deployment.NewDeployment(dep, time.Second*30)

	tlsObj := &TLS{
		Service: &Service{
			SecretName: "tls-secret-name",
		},
		Ca: &Ca{
			CaSecretName: "ca-secret-name",
		},
	}

	logger := log.Log.WithName("test-logger")

	helperObj, err := helper.NewHelper(dep, k8sClient, nil, k8sClient.Scheme(), logger)
	if err != nil {
		t.Fatalf("failed to create helper: %v", err)
	}

	err = tlsObj.UpdateDeploymentWithTLS(context.Background(), customDeployment, helperObj)
	assert.Nil(err, "failed to update deployment with TLS")

	updatedDep := &appsv1.Deployment{}
	err = k8sClient.Get(context.Background(), client.ObjectKey{Name: "test-deployment", Namespace: "default"}, updatedDep)
	assert.Nil(err, "failed to get updated deployment")

	assert.NotZero(len(updatedDep.Spec.Template.Spec.Volumes), "expected TLS volumes to be added but found none")
}
