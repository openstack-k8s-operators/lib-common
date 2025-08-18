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

// Package annotations provides utilities for managing network annotations and network attachment definitions
package annotations

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// NetworkAttachmentAnnot pod annotation for network-attachment-definition
// (mschuppert) for now specify the const from "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
// here to not have that dependency. If there is more we get from there could import it
const NetworkAttachmentAnnot string = "k8s.v1.cni.cncf.io/networks"

type net struct {
	Name      string
	Namespace string
}

// GetNADAnnotation returns pod annotation for network-attachment-definition
// e.g. k8s.v1.cni.cncf.io/networks: '[{"name": "internalapi", "namespace": "openstack"},{"name": "storage", "namespace": "openstack"}]'
// DEPRECATED in favor of CreateNetworksAnnotation from network pkg
func GetNADAnnotation(namespace string, nads []string) (map[string]string, error) {

	netAnnotations := []net{}
	for _, nad := range nads {
		netAnnotations = append(
			netAnnotations,
			net{
				Name:      nad,
				Namespace: namespace,
			},
		)
	}

	networks, err := json.Marshal(netAnnotations)
	if err != nil {
		return nil, fmt.Errorf("failed to encode networks %s into json: %w", nads, err)
	}

	return map[string]string{NetworkAttachmentAnnot: string(networks)}, nil
}

// GetBoolFromAnnotation - it returns a boolean from a string annotation
// e.g. glance.openstack.org/wsgi: "true" returns true as a boolean type
//
// Cases covered by this function:
// 1. the annotation does not exist -> false, false, nil
// 2. the annotation exist and is not a valid boolean -> false, true, error
// 3. the annotation exists and is a valid False bool -> false, true, nil
// 4. the annotation exists and is a valid True bool -> true, true, nil
func GetBoolFromAnnotation(
	ann map[string]string,
	key string,
) (bool, bool, error) {
	// Try to get the value associated to the annotation key
	value, exists := ann[key]
	if !exists {
		return false, false, nil
	}
	result, err := strconv.ParseBool(value)
	if err != nil {
		// the annotation is not a valid boolean, return an error
		return false, exists, err
	}
	return result, exists, nil
}
