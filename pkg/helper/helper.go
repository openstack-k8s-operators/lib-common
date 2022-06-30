/*
Copyright 2020 Red Hat

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

package helper

import (
	"encoding/json"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

// Helper is a utility for ensuring the proper patching of objects.
type Helper struct {
	client       client.Client
	kclient      kubernetes.Interface
	gvk          schema.GroupVersionKind
	scheme       *runtime.Scheme
	beforeObject client.Object
	before       *unstructured.Unstructured
	after        *unstructured.Unstructured
	changes      map[string]bool
	finalizer    string

	logger logr.Logger
}

// NewHelper returns an initialized Helper.
func NewHelper(obj client.Object, crClient client.Client, kclient kubernetes.Interface, scheme *runtime.Scheme, log logr.Logger) (*Helper, error) {
	// Get the GroupVersionKind of the object,
	// used to validate against later on.
	gvk, err := apiutil.GVKForObject(obj, crClient.Scheme())
	if err != nil {
		return nil, err
	}

	// Convert the object to unstructured to compare against our before copy.
	unstructuredObj, err := ToUnstructured(obj)
	if err != nil {
		return nil, err
	}

	return &Helper{
		client:       crClient,
		kclient:      kclient,
		gvk:          gvk,
		scheme:       scheme,
		before:       unstructuredObj,
		beforeObject: obj.DeepCopyObject().(client.Object),
		logger:       log,
		finalizer:    gvk.Kind,
	}, nil
}

// GetClient - returns the client
func (h *Helper) GetClient() client.Client {
	return h.client
}

// GetKClient - returns the kclient
func (h *Helper) GetKClient() kubernetes.Interface {
	return h.kclient
}

// GetGKV - returns the GKV of the object
func (h *Helper) GetGKV() schema.GroupVersionKind {
	return h.gvk
}

// GetScheme - returns the runtime scheme of the object
func (h *Helper) GetScheme() *runtime.Scheme {
	return h.scheme
}

// GetAfter - returns unstructured object after modification
func (h *Helper) GetAfter() *unstructured.Unstructured {
	return h.after
}

// GetBefore - returns unstructured object after modification
func (h *Helper) GetBefore() *unstructured.Unstructured {
	return h.before
}

// GetChanges - returns unstructured object after modification
func (h *Helper) GetChanges() map[string]bool {
	return h.changes
}

// GetBeforeObject - returns the object before modification
func (h *Helper) GetBeforeObject() client.Object {
	return h.beforeObject
}

// GetLogger - returns the logger
func (h *Helper) GetLogger() logr.Logger {
	return h.logger
}

// GetFinalizer - returns the finalizer
func (h *Helper) GetFinalizer() string {
	return h.finalizer
}

// SetAfter - returns the logger
func (h *Helper) SetAfter(obj client.Object) error {
	unstructuredObj, err := ToUnstructured(obj)
	if err != nil {
		return err
	}

	h.after = unstructuredObj

	// Calculate and store the top-level field changes (e.g. "metadata", "spec", "status") we have before/after.
	h.changes, err = h.calculateChanges(obj)
	if err != nil {
		return err
	}

	return nil
}

// calculateChanges - calculate changes tries to build a patch from the before/after objects we have
// and store in a map which top-level fields (e.g. `metadata`, `spec`, `status`, etc.) have changed.
func (h *Helper) calculateChanges(after client.Object) (map[string]bool, error) {
	// Calculate patch data.
	patch := client.MergeFrom(h.beforeObject)
	diff, err := patch.Data(after)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to calculate patch data")
	}

	// Unmarshal patch data into a local map.
	patchDiff := map[string]interface{}{}
	if err := json.Unmarshal(diff, &patchDiff); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal patch data into a map")
	}

	// Return the map.
	res := make(map[string]bool, len(patchDiff))
	for key := range patchDiff {
		res[key] = true
	}
	return res, nil
}

// ToUnstructured - convert to unstructured
func ToUnstructured(obj runtime.Object) (*unstructured.Unstructured, error) {
	// If the incoming object is already unstructured, perform a deep copy first
	// otherwise DefaultUnstructuredConverter ends up returning the inner map without
	// making a copy.
	if _, ok := obj.(runtime.Unstructured); ok {
		obj = obj.DeepCopyObject()
	}
	rawMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}
	return &unstructured.Unstructured{Object: rawMap}, nil
}
