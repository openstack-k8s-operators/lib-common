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

package backup

import (
	"context"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Config holds backup/restore configuration for a CRD
type Config struct {
	Enabled      bool
	RestoreOrder string
	Category     string
}

// CRDLabelCache maps CRD names to their backup configuration
type CRDLabelCache map[string]Config

// BuildCRDLabelCache reads all CRDs and caches their backup labels
func BuildCRDLabelCache(ctx context.Context, c client.Client) (CRDLabelCache, error) {
	cache := make(CRDLabelCache)

	crdList := &apiextensionsv1.CustomResourceDefinitionList{}
	if err := c.List(ctx, crdList); err != nil {
		return nil, err
	}

	for _, crd := range crdList.Items {
		labels := crd.GetLabels()
		if labels == nil {
			continue
		}

		// Only cache CRDs that opt into backup/restore
		if labels[BackupRestoreLabel] != "true" {
			continue
		}

		config := Config{
			Enabled:      true,
			RestoreOrder: labels[BackupRestoreOrderLabel],
			Category:     labels[BackupCategoryLabel],
		}

		// Cache by CRD name (e.g., "keystoneapis.keystone.openstack.org")
		cache[crd.Name] = config
	}

	return cache, nil
}

// GetConfig looks up backup configuration by CRD name
// (e.g., "keystoneapis.keystone.openstack.org").
// Returns Config with Enabled=false if not found.
func (c CRDLabelCache) GetConfig(crdName string) Config {
	if config, ok := c[crdName]; ok {
		return config
	}
	return Config{Enabled: false}
}
