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

package annotations

import (
	"encoding/json"
	"fmt"
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
