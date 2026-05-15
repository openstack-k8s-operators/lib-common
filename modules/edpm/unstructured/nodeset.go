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

package unstructured

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	k8s_unstructured "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	oko_secret "github.com/openstack-k8s-operators/lib-common/modules/common/secret"
)

// NodeSetGVK is the GroupVersionKind for OpenStackDataPlaneNodeSet.
// Exported so callers can use it for controller watches on unstructured objects.
var NodeSetGVK = schema.GroupVersionKind{
	Group:   "dataplane.openstack.org",
	Version: "v1beta1",
	Kind:    "OpenStackDataPlaneNodeSet",
}

// NewNodeSetObject returns an unstructured OpenStackDataPlaneNodeSet object,
// suitable for use with controller-runtime Watches.
func NewNodeSetObject() *k8s_unstructured.Unstructured {
	obj := &k8s_unstructured.Unstructured{}
	obj.SetGroupVersionKind(NodeSetGVK)
	return obj
}

// AreSecretHashesInSync checks whether the deployed secret hashes in all
// OpenStackDataPlaneNodeSets in the given namespace match the current cluster
// secrets. It uses an unstructured client to avoid importing dataplane API types.
//
// Returns:
//   - inSync=true when all hashes match, no nodesets exist, or the
//     OpenStackDataPlaneNodeSet CRD is not installed on the cluster.
//   - inSync=false with info describing the first mismatch when a secret
//     has changed since the last full deployment or has been deleted.
func AreSecretHashesInSync(
	ctx context.Context,
	c client.Client,
	namespace string,
) (inSync bool, info string, err error) {
	Log := log.FromContext(ctx)

	nodesetList := &k8s_unstructured.UnstructuredList{}
	nodesetList.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   NodeSetGVK.Group,
		Version: NodeSetGVK.Version,
		Kind:    NodeSetGVK.Kind + "List",
	})

	if err := c.List(ctx, nodesetList, client.InNamespace(namespace)); err != nil {
		if meta.IsNoMatchError(err) {
			Log.Info("OpenStackDataPlaneNodeSet CRD not installed, skipping hash check")
			return true, "", nil
		}
		return false, "", fmt.Errorf("failed to list OpenStackDataPlaneNodeSets: %w", err)
	}

	if len(nodesetList.Items) == 0 {
		Log.Info("No nodesets found in namespace - secrets in sync",
			"namespace", namespace)
		return true, "", nil
	}

	for i := range nodesetList.Items {
		item := &nodesetList.Items[i]

		if err := ctx.Err(); err != nil {
			return false, "", fmt.Errorf("context cancelled during nodeset check: %w", err)
		}

		secretHashes, found, err := k8s_unstructured.NestedStringMap(item.Object, "status", "secretHashes")
		if err != nil {
			return false, "", fmt.Errorf("failed to read secretHashes from nodeset %s/%s: %w",
				item.GetNamespace(), item.GetName(), err)
		}
		if !found || len(secretHashes) == 0 {
			continue
		}

		for secretName, deployedHash := range secretHashes {
			currentSecret := &corev1.Secret{}
			err := c.Get(ctx, types.NamespacedName{
				Name:      secretName,
				Namespace: namespace,
			}, currentSecret)
			if err != nil {
				if k8s_errors.IsNotFound(err) {
					info := fmt.Sprintf("nodeset %s/%s: deployed secret %s no longer exists",
						item.GetNamespace(), item.GetName(), secretName)
					return false, info, nil
				}
				return false, "", fmt.Errorf("failed to get secret %s: %w", secretName, err)
			}

			currentHash, hashErr := oko_secret.Hash(currentSecret)
			if hashErr != nil {
				return false, "", fmt.Errorf("failed to hash secret %s: %w", secretName, hashErr)
			}

			if currentHash != deployedHash {
				info := fmt.Sprintf("nodeset %s/%s: secret %s has changed since last full deployment",
					item.GetNamespace(), item.GetName(), secretName)
				return false, info, nil
			}
		}
	}

	Log.Info("All nodeset secret hashes match - secrets in sync",
		"namespace", namespace, "nodesetsChecked", len(nodesetList.Items))
	return true, "", nil
}

// HaveNodeSets returns true if any OpenStackDataPlaneNodeSets with non-empty
// status.secretHashes exist in the given namespace. Returns false when no
// NodeSets exist, none have secretHashes, or the CRD is not installed.
func HaveNodeSets(ctx context.Context, c client.Client, namespace string) (bool, error) {
	nodesetList := &k8s_unstructured.UnstructuredList{}
	nodesetList.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   NodeSetGVK.Group,
		Version: NodeSetGVK.Version,
		Kind:    NodeSetGVK.Kind + "List",
	})

	if err := c.List(ctx, nodesetList, client.InNamespace(namespace)); err != nil {
		if meta.IsNoMatchError(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to list OpenStackDataPlaneNodeSets: %w", err)
	}

	for i := range nodesetList.Items {
		item := &nodesetList.Items[i]
		secretHashes, found, err := k8s_unstructured.NestedStringMap(item.Object, "status", "secretHashes")
		if err != nil {
			return false, fmt.Errorf("failed to read secretHashes from nodeset %s: %w", item.GetName(), err)
		}
		if found && len(secretHashes) > 0 {
			return true, nil
		}
	}

	return false, nil
}
