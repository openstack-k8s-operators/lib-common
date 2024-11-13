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

package networkattachment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"

	networkv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	"github.com/openstack-k8s-operators/lib-common/modules/common/helper"
	"github.com/openstack-k8s-operators/lib-common/modules/common/pod"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/jsonpath"
)

// GetNADWithName - Get network-attachment-definition with name in namespace
func GetNADWithName(
	ctx context.Context,
	h *helper.Helper,
	name string,
	namespace string,
) (*networkv1.NetworkAttachmentDefinition, error) {

	nad := &networkv1.NetworkAttachmentDefinition{}

	err := h.GetClient().Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, nad)
	if err != nil {
		err = fmt.Errorf("Error getting network-attachment-definition %s/%s - %w", name, namespace, err)

		return nil, err
	}

	return nad, nil
}

// CreateNetworksAnnotation returns pod annotation for network-attachment-definition list
// e.g. k8s.v1.cni.cncf.io/networks: '[{"name": "internalapi", "namespace": "openstack"},{"name": "storage", "namespace": "openstack"}]'
// NOTE: Deprecated, use EnsureNetworksAnnotation
func CreateNetworksAnnotation(namespace string, nads []string) (map[string]string, error) {

	netAnnotations := []networkv1.NetworkSelectionElement{}
	for _, nad := range nads {
		netAnnotations = append(
			netAnnotations,
			networkv1.NetworkSelectionElement{
				Name:             nad,
				Namespace:        namespace,
				InterfaceRequest: GetNetworkIFName(nad),
			},
		)
	}

	networks, err := json.Marshal(netAnnotations)
	if err != nil {
		return nil, fmt.Errorf("failed to encode networks %s into json: %w", nads, err)
	}

	return map[string]string{networkv1.NetworkAttachmentAnnot: string(networks)}, nil
}

// GetNetworkIFName returns the interface name base on the NAD name
// the interface name in Linux must not be longer then 15 chars.
func GetNetworkIFName(nad string) string {
	if len(nad) > 15 {
		return nad[:15]
	}
	return nad
}

// GetNetworkStatusFromAnnotation returns NetworkStatus list with networking details the pods are attached to
func GetNetworkStatusFromAnnotation(annotations map[string]string) ([]networkv1.NetworkStatus, error) {

	var netStatus []networkv1.NetworkStatus

	if netStatusAnnotation, ok := annotations[networkv1.NetworkStatusAnnot]; ok {
		err := json.Unmarshal([]byte(netStatusAnnotation), &netStatus)
		if err != nil {
			return nil, fmt.Errorf("failed to decode networks status %s: %w", netStatusAnnotation, err)
		}
	}

	return netStatus, nil
}

// VerifyNetworkStatusFromAnnotation - verifies if NetworkStatus annotation for the pods of a deployment,
// pods identified via the service label selector, matches the passed in network attachments and the number of
// per network IPs the ready count of the deployment. Return true if count matches with the list of IPs per network.
func VerifyNetworkStatusFromAnnotation(
	ctx context.Context,
	helper *helper.Helper,
	networkAttachments []string,
	serviceLabels map[string]string,
	readyCount int32,
) (bool, map[string][]string, error) {

	networkReady := true
	networkAttachmentStatus := map[string][]string{}
	if len(networkAttachments) > 0 {
		podList, err := pod.GetPodListWithLabel(ctx, helper, helper.GetBeforeObject().GetNamespace(), serviceLabels)
		if err != nil {
			return false, networkAttachmentStatus, err
		}

		for _, pod := range podList.Items {
			netsStatus, err := GetNetworkStatusFromAnnotation(pod.Annotations)
			if err != nil {
				return networkReady, networkAttachmentStatus, err
			}
			for _, netStat := range netsStatus {
				networkAttachmentStatus[netStat.Name] = append(networkAttachmentStatus[netStat.Name], netStat.IPs...)
			}
		}

		for _, netAtt := range networkAttachments {
			netAtt = helper.GetBeforeObject().GetNamespace() + "/" + netAtt

			if net, ok := networkAttachmentStatus[netAtt]; !ok || len(net) < int(readyCount) {
				networkReady = false
				break
			}
		}
	}

	return networkReady, networkAttachmentStatus, nil
}

// EnsureNetworksAnnotation returns pod annotation for network-attachment-definition list
// e.g. k8s.v1.cni.cncf.io/networks: '[{"name": "internalapi", "namespace": "openstack"},{"name": "storage", "namespace": "openstack"}]'
// If `ipam.gateway` is defined in the NAD, the annotation will contain the `default-route` for that network:
// e.g. k8s.v1.cni.cncf.io/networks: '[{"name":"internalapi","namespace":"openstack","interface":"internalapi","default-route":["10.1.2.200"]}]'
func EnsureNetworksAnnotation(
	nadList []networkv1.NetworkAttachmentDefinition,
) (map[string]string, error) {

	annotationString := map[string]string{}
	netAnnotations := []networkv1.NetworkSelectionElement{}
	for _, nad := range nadList {
		gateway := ""

		var data interface{}
		if err := json.Unmarshal([]byte(nad.Spec.Config), &data); err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON data: %w", err)
		}

		// use jsonpath to parse the cni config
		jp := jsonpath.New(nad.Name)
		jp.AllowMissingKeys(true) // Allow missing keys, for when no gateway configured

		// Parse the JSONPath template, for now just `ipam.gateway`
		err := jp.Parse(`{.ipam.gateway}`)
		if err != nil {
			return annotationString, fmt.Errorf("parse template error: %w", err)
		}

		buf := new(bytes.Buffer)
		// get the gateway from the config
		err = jp.Execute(buf, data)
		if err != nil {
			return annotationString, fmt.Errorf("parse execute template against nad %+v error: %w", nad.Spec.Config, err)
		}

		gateway = buf.String()

		gatewayReq := []net.IP{}
		if gateway != "" {
			gatewayReq = append(gatewayReq, net.ParseIP(gateway))

		}

		netAnnotations = append(
			netAnnotations,
			networkv1.NetworkSelectionElement{
				Name:             nad.Name,
				Namespace:        nad.Namespace,
				InterfaceRequest: GetNetworkIFName(nad.Name),
				GatewayRequest:   gatewayReq,
			},
		)
	}

	networks, err := json.Marshal(netAnnotations)
	if err != nil {
		return nil, fmt.Errorf("failed to encode networks %v into json: %w", nadList, err)
	}

	annotationString[networkv1.NetworkAttachmentAnnot] = string(networks)

	return annotationString, nil
}
