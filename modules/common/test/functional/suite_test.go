/*
Copyright 2023.

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

package functional

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	"go.uber.org/zap/zapcore"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	routev1 "github.com/openshift/api/route/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	. "github.com/openstack-k8s-operators/lib-common/modules/common/test/helpers"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

const (
	timeout = 10 * time.Second
	// have maximum 100 retries before the timeout hits
	interval = timeout / 100
	// consistencyTimeout is the amount of time we use to repeatedly check
	// that a condition is still valid. This is intended to be used in
	// asserts using `Consistently`.
	// consistencyTimeout = timeout
)

var (
	cfg     *rest.Config
	cClient client.Client
	testEnv *envtest.Environment
	ctx     context.Context
	cancel  context.CancelFunc
	logger  logr.Logger
	h       *helper.Helper
	th      *TestHelper
)

func TestCommon(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "common module suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true), func(o *zap.Options) {
		o.Development = true
		o.TimeEncoder = zapcore.ISO8601TimeEncoder
	}))

	ctx, cancel = context.WithCancel(context.TODO())

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("..", "..", "..", "test", "openshift_crds", "route", "v1"),
		},
		ErrorIfCRDPathMissing: true,
	}
	var err error

	// cfg is defined in this file globally.
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = batchv1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	err = corev1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	err = routev1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	//+kubebuilder:scaffold:scheme

	logger = ctrl.Log.WithName("---Test---")

	cClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(cClient).NotTo(BeNil())

	client, err := kubernetes.NewForConfig(cfg)
	Expect(err).NotTo(HaveOccurred())
	Expect(client).NotTo(BeNil())

	th = NewTestHelper(ctx, cClient, timeout, interval, logger)
	Expect(th).NotTo(BeNil())

	// NOTE(gibi): helper.Helper needs an object that is being reconciled
	// we are not really doing reconciliation in this test but still we need to
	// provide a valid object. It is used as controller reference for certain
	// objects created in the test. So we provide a simple one, a Namespace.
	genericObject := th.CreateNamespace("generic-object")
	h, err = helper.NewHelper(genericObject, cClient, client, testEnv.Scheme, ctrl.Log)
	Expect(err).NotTo(HaveOccurred())
	Expect(h).NotTo(BeNil())

	go func() {
		defer GinkgoRecover()
	}()

})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	cancel()
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
