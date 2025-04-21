/*
Copyright 2022 Red Hat
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
	"context"
	"os"
	"strconv"
	"time"

	"github.com/onsi/gomega"

	"github.com/go-logr/logr"
	"github.com/openstack-k8s-operators/lib-common/modules/common/condition"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	memcachedv1 "github.com/openstack-k8s-operators/infra-operator/apis/memcached/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
)

type conditionsGetter interface {
	GetConditions(name types.NamespacedName) condition.Conditions
}

// ConditionGetterFunc recieves custom condition getters for operators specific needs
type ConditionGetterFunc func(name types.NamespacedName) condition.Conditions

// GetConditions implements conditions getter for operators specific needs
func (f ConditionGetterFunc) GetConditions(name types.NamespacedName) condition.Conditions {
	return f(name)
}

// TestHelper is a collection EnvTest helpers writing test to code that
// interacts with the k8s resource. If you need to handle openstack-k8s-operators
// specific resource then you should use the extended TestHelper from
// modules/test-operator
type TestHelper struct {
	K8sClient client.Client
	Ctx       context.Context
	Timeout   time.Duration
	Interval  time.Duration
	Logger    logr.Logger
}

// NewTestHelper returns a TestHelper
func NewTestHelper(
	ctx context.Context,
	k8sClient client.Client,
	timeout time.Duration,
	interval time.Duration,
	logger logr.Logger,
) *TestHelper {
	return &TestHelper{
		Ctx:       ctx,
		K8sClient: k8sClient,
		Timeout:   getTestTimeout(timeout),
		Interval:  interval,
		Logger:    logger,
	}
}

// getTestTimeout returns test timeout from TEST_TIMEOUT_SEC environment
// variable, in seconds; or picks defaultTimeout, in milliseconds
func getTestTimeout(defaultTimeout time.Duration) time.Duration {
	t := os.Getenv("TEST_TIMEOUT_SEC")
	timeout, err := strconv.Atoi(t)
	if err != nil {
		return defaultTimeout
	}
	return time.Duration(timeout) * time.Second
}

// CreateUnstructured creates an unstructured Kubernetes object from a map of key-value pairs.
//
// Example usage:
//
//	  rawObj := map[string]interface{}{
//	    "apiVersion": "nova.openstack.org/v1beta1",
//			"kind":       "NovaAPI",
//			"metadata": map[string]interface{}{
//				"name":      name.Name,
//				"namespace": name.Namespace,
//			},
//			"spec": spec,
//	    },
//	    ...
//	  }
//	  unstructuredObj := tc.CreateUnstructured(rawObj)
func (tc *TestHelper) CreateUnstructured(rawObj map[string]interface{}) *unstructured.Unstructured {
	tc.Logger.Info("Creating", "raw", rawObj)
	unstructuredObj := &unstructured.Unstructured{Object: rawObj}
	_, err := controllerutil.CreateOrPatch(
		tc.Ctx, tc.K8sClient, unstructuredObj, func() error { return nil })
	gomega.Expect(err).ShouldNot(gomega.HaveOccurred())

	return unstructuredObj
}

// GetName function is used only in lib-common
func (tc *TestHelper) GetName(obj client.Object) types.NamespacedName {
	return types.NamespacedName{Namespace: obj.GetNamespace(), Name: obj.GetName()}
}

// GetEnvVarValue returns the value of the EnvVar based on the name of the Var
// or return the defaultValue if the list does not have EnvVar with the given name
func GetEnvVarValue(envs []corev1.EnvVar, name string, defaultValue string) string {
	for _, e := range envs {
		if e.Name == name {
			return e.Value
		}
	}
	return defaultValue
}

// GetDefaultMemcachedSpec return memcachedv1.MemcachedSpec for test-helpers in operators
func GetDefaultMemcachedSpec() memcachedv1.MemcachedSpec {
	return memcachedv1.MemcachedSpec{
		MemcachedSpecCore: memcachedv1.MemcachedSpecCore{
			Replicas: ptr.To(int32(3)),
		},
	}
}
