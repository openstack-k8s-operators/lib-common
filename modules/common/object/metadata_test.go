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

package object

import (
	"context"
	"errors"
	"testing"
	"time"

	commonannotations "github.com/openstack-k8s-operators/lib-common/modules/common/annotations"
	"github.com/openstack-k8s-operators/lib-common/modules/common/helper"

	. "github.com/onsi/gomega" // nolint:revive

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var errSyntheticAPI = errors.New("synthetic API server error")

var (
	metadata = metav1.ObjectMeta{
		Name:      "foo",
		Namespace: "bar",
		OwnerReferences: []metav1.OwnerReference{
			{
				APIVersion:         "core.openstack.org/v1beta1",
				BlockOwnerDeletion: ptr.To(true),
				Controller:         ptr.To(true),
				Kind:               "OpenStackControlPlane",
				Name:               "openstack-network-isolation",
				UID:                "11111111-1111-1111-1111-111111111111",
			},
		},
	}
)

func setupHelper(objs ...client.Object) (*helper.Helper, error) {
	s := scheme.Scheme

	fakeClient := fake.NewClientBuilder().
		WithScheme(s).
		WithObjects(objs...).
		Build()

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: "test-namespace"},
	}

	return helper.NewHelper(ns, fakeClient, nil, s, ctrl.Log)
}

func setupHelperWithInterceptor(funcs interceptor.Funcs, objs ...client.Object) (*helper.Helper, error) {
	s := scheme.Scheme

	fakeClient := fake.NewClientBuilder().
		WithScheme(s).
		WithObjects(objs...).
		WithInterceptorFuncs(funcs).
		Build()

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: "test-namespace"},
	}

	return helper.NewHelper(ns, fakeClient, nil, s, ctrl.Log)
}

func TestCheckOwnerRefExist(t *testing.T) {
	tests := []struct {
		name      string
		ownerRefs []metav1.OwnerReference
		uid       types.UID
		want      bool
	}{
		{
			name:      "Check existing owner",
			ownerRefs: metadata.OwnerReferences,
			uid:       types.UID("11111111-1111-1111-1111-111111111111"),
			want:      true,
		},
		{
			name:      "Check non existing owner",
			ownerRefs: metadata.OwnerReferences,
			uid:       types.UID("22222222-2222-2222-2222-222222222222"),
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			g.Expect(CheckOwnerRefExist(tt.uid, tt.ownerRefs)).To(BeIdenticalTo(tt.want))
		})
	}
}

func TestFinalizeSecretRotation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	const (
		namespace = "openstack"
		finalizer = "openstack.org/test-consumer"
	)

	t.Run("early return paths", func(t *testing.T) {
		tests := []struct {
			name          string
			statusSecret  string
			currentSecret string
			guardReady    bool
			wantSecret    string
		}{
			{
				name:          "no rotation - empty status",
				statusSecret:  "",
				currentSecret: "new-secret",
				guardReady:    true,
				wantSecret:    "new-secret",
			},
			{
				name:          "no rotation - same secret",
				statusSecret:  "same-secret",
				currentSecret: "same-secret",
				guardReady:    true,
				wantSecret:    "same-secret",
			},
			{
				name:          "rotation in progress - guard not ready",
				statusSecret:  "old-secret",
				currentSecret: "new-secret",
				guardReady:    false,
				wantSecret:    "old-secret",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				g := NewWithT(t)

				got, err := FinalizeSecretRotation(
					ctx, nil, namespace,
					tt.statusSecret, tt.currentSecret,
					finalizer, tt.guardReady,
				)

				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(got).To(Equal(tt.wantSecret))
			})
		}
	})

	t.Run("rotation complete - removes finalizer from old secret", func(t *testing.T) {
		g := NewWithT(t)

		oldSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "old-secret",
				Namespace: namespace,
			},
		}
		controllerutil.AddFinalizer(oldSecret, finalizer)

		h, err := setupHelper(oldSecret)
		g.Expect(err).NotTo(HaveOccurred())

		got, err := FinalizeSecretRotation(
			ctx, h, namespace,
			"old-secret", "new-secret",
			finalizer, true,
		)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(got).To(Equal("new-secret"))

		updated := &corev1.Secret{}
		g.Expect(h.GetClient().Get(ctx, types.NamespacedName{
			Name: "old-secret", Namespace: namespace,
		}, updated)).To(Succeed())
		g.Expect(controllerutil.ContainsFinalizer(updated, finalizer)).To(BeFalse())
	})

	t.Run("rotation complete - old secret already deleted", func(t *testing.T) {
		g := NewWithT(t)

		h, err := setupHelper()
		g.Expect(err).NotTo(HaveOccurred())

		got, err := FinalizeSecretRotation(
			ctx, h, namespace,
			"old-secret", "new-secret",
			finalizer, true,
		)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(got).To(Equal("new-secret"))
	})

	t.Run("rotation complete - Get fails with non-NotFound error", func(t *testing.T) {
		g := NewWithT(t)

		h, err := setupHelperWithInterceptor(interceptor.Funcs{
			Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
				if _, ok := obj.(*corev1.Secret); ok {
					return errSyntheticAPI
				}
				return c.Get(ctx, key, obj, opts...)
			},
		})
		g.Expect(err).NotTo(HaveOccurred())

		got, err := FinalizeSecretRotation(
			ctx, h, namespace,
			"old-secret", "new-secret",
			finalizer, true,
		)
		g.Expect(err).To(HaveOccurred())
		g.Expect(got).To(Equal("old-secret"))
	})
}

func TestManageRotationGracePeriod(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	newConfigMap := func(annotations map[string]string) *corev1.ConfigMap {
		return &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "test-instance",
				Namespace:   "openstack",
				Annotations: annotations,
			},
		}
	}

	newFakeClient := func(objs ...client.Object) client.Client {
		return fake.NewClientBuilder().
			WithScheme(scheme.Scheme).
			WithObjects(objs...).
			Build()
	}

	getAnnotation := func(g Gomega, c client.Client, cm *corev1.ConfigMap) string {
		updated := &corev1.ConfigMap{}
		g.Expect(c.Get(ctx, types.NamespacedName{
			Name: cm.Name, Namespace: cm.Namespace,
		}, updated)).To(Succeed())
		return updated.Annotations[commonannotations.RotationGraceAnnotation]
	}

	t.Run("not pending, no annotation - no-op", func(t *testing.T) {
		g := NewWithT(t)
		cm := newConfigMap(nil)
		c := newFakeClient(cm)

		result, graceActive, err := ManageRotationGracePeriod(ctx, c, cm, false, 30*time.Second)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(graceActive).To(BeFalse())
		g.Expect(result).To(Equal(ctrl.Result{}))
	})

	t.Run("not pending, annotation present - clears it", func(t *testing.T) {
		g := NewWithT(t)
		cm := newConfigMap(map[string]string{
			commonannotations.RotationGraceAnnotation: time.Now().Add(1 * time.Minute).Format(time.RFC3339),
		})
		c := newFakeClient(cm)

		result, graceActive, err := ManageRotationGracePeriod(ctx, c, cm, false, 30*time.Second)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(graceActive).To(BeFalse())
		g.Expect(result).To(Equal(ctrl.Result{}))
		g.Expect(getAnnotation(g, c, cm)).To(BeEmpty())
	})

	t.Run("pending, no annotation - sets grace period", func(t *testing.T) {
		g := NewWithT(t)
		cm := newConfigMap(nil)
		c := newFakeClient(cm)

		before := time.Now()
		result, graceActive, err := ManageRotationGracePeriod(ctx, c, cm, true, 30*time.Second)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(graceActive).To(BeTrue())
		g.Expect(result.RequeueAfter).To(Equal(30 * time.Second))

		ann := getAnnotation(g, c, cm)
		g.Expect(ann).NotTo(BeEmpty())
		parsed, parseErr := time.Parse(time.RFC3339, ann)
		g.Expect(parseErr).NotTo(HaveOccurred())
		g.Expect(parsed).To(BeTemporally("~", before.Add(30*time.Second), 2*time.Second))
	})

	t.Run("pending, annotation in future - requeues with remaining", func(t *testing.T) {
		g := NewWithT(t)
		future := time.Now().Add(1 * time.Minute)
		cm := newConfigMap(map[string]string{
			commonannotations.RotationGraceAnnotation: future.Format(time.RFC3339),
		})
		c := newFakeClient(cm)

		result, graceActive, err := ManageRotationGracePeriod(ctx, c, cm, true, 30*time.Second)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(graceActive).To(BeTrue())
		g.Expect(result.RequeueAfter).To(BeNumerically(">", 0))
		g.Expect(result.RequeueAfter).To(BeNumerically("<=", 1*time.Minute))
	})

	t.Run("pending, annotation expired - grace period over", func(t *testing.T) {
		g := NewWithT(t)
		past := time.Now().Add(-1 * time.Minute)
		cm := newConfigMap(map[string]string{
			commonannotations.RotationGraceAnnotation: past.Format(time.RFC3339),
		})
		c := newFakeClient(cm)

		result, graceActive, err := ManageRotationGracePeriod(ctx, c, cm, true, 30*time.Second)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(graceActive).To(BeFalse())
		g.Expect(result).To(Equal(ctrl.Result{}))
	})

	t.Run("pending, malformed annotation - clears it", func(t *testing.T) {
		g := NewWithT(t)
		cm := newConfigMap(map[string]string{
			commonannotations.RotationGraceAnnotation: "not-a-timestamp",
		})
		c := newFakeClient(cm)

		result, graceActive, err := ManageRotationGracePeriod(ctx, c, cm, true, 30*time.Second)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(graceActive).To(BeFalse())
		g.Expect(result).To(Equal(ctrl.Result{}))
		g.Expect(getAnnotation(g, c, cm)).To(BeEmpty())
	})
}
