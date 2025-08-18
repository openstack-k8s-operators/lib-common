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
	"testing"

	networkv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/gomega" // nolint:revive
)

func TestCreateNetworksAnnotation(t *testing.T) {

	tests := []struct {
		name      string
		networks  []string
		namespace string
		want      map[string]string
	}{
		{
			name:      "No network",
			networks:  []string{},
			namespace: "foo",
			want:      map[string]string{networkv1.NetworkAttachmentAnnot: "[]"},
		},
		{
			name:      "Single network",
			networks:  []string{"one"},
			namespace: "foo",
			want:      map[string]string{networkv1.NetworkAttachmentAnnot: "[{\"name\":\"one\",\"namespace\":\"foo\",\"interface\":\"one\"}]"},
		},
		{
			name:      "Multiple networks",
			networks:  []string{"one", "two"},
			namespace: "foo",
			want:      map[string]string{networkv1.NetworkAttachmentAnnot: "[{\"name\":\"one\",\"namespace\":\"foo\",\"interface\":\"one\"},{\"name\":\"two\",\"namespace\":\"foo\",\"interface\":\"two\"}]"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			networkAnnotation, err := CreateNetworksAnnotation(tt.namespace, tt.networks)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(networkAnnotation).To(HaveLen(len(tt.want)))
			g.Expect(networkAnnotation).To(BeEquivalentTo(tt.want))
		})
	}
}

func TestGetNetworkStatusFromAnnotation(t *testing.T) {

	tests := []struct {
		name        string
		annotations map[string]string
		want        []networkv1.NetworkStatus
	}{
		{
			name:        "Empty annotation",
			annotations: map[string]string{},
			want:        nil,
		},
		{
			name: "just pod network",
			annotations: map[string]string{
				"k8s.v1.cni.cncf.io/network-status":  "[{\n    \"name\": \"openshift-sdn\",\n    \"interface\": \"eth0\",\n    \"ips\": [\n        \"10.131.0.16\"\n    ],\n    \"default\": true,\n    \"dns\": {}\n}]",
				"k8s.v1.cni.cncf.io/networks-status": "[{\n    \"name\": \"openshift-sdn\",\n    \"interface\": \"eth0\",\n    \"ips\": [\n        \"10.131.0.16\"\n    ],\n    \"default\": true,\n    \"dns\": {}\n}]",
			},
			want: []networkv1.NetworkStatus{
				{
					Name:       "openshift-sdn",
					Interface:  "eth0",
					IPs:        []string{"10.131.0.16"},
					Mac:        "",
					Default:    true,
					DNS:        networkv1.DNS{Nameservers: nil, Domain: "", Search: nil, Options: nil},
					DeviceInfo: nil,
					Gateway:    nil,
				},
			},
		},
		{
			name: "with additional networkAttachment",
			annotations: map[string]string{
				"k8s.v1.cni.cncf.io/network-status":  "[{\n    \"name\": \"openshift-sdn\",\n    \"interface\": \"eth0\",\n    \"ips\": [\n        \"10.130.0.16\"\n    ],\n    \"default\": true,\n    \"dns\": {}\n},{\n    \"name\": \"openstack/internalapi\",\n    \"interface\": \"net1\",\n    \"ips\": [\n        \"172.17.0.226\"\n    ],\n    \"mac\": \"a2:ef:bb:ae:65:45\",\n    \"dns\": {}\n}]",
				"k8s.v1.cni.cncf.io/networks":        "[{\"name\":\"internalapi\",\"namespace\":\"openstack\"}]",
				"k8s.v1.cni.cncf.io/networks-status": "[{\n    \"name\": \"openshift-sdn\",\n    \"interface\": \"eth0\",\n    \"ips\": [\n        \"10.130.0.16\"\n    ],\n    \"default\": true,\n    \"dns\": {}\n},{\n    \"name\": \"openstack/internalapi\",\n    \"interface\": \"net1\",\n    \"ips\": [\n        \"172.17.0.226\"\n    ],\n    \"mac\": \"a2:ef:bb:ae:65:45\",\n    \"dns\": {}\n}]",
			},
			want: []networkv1.NetworkStatus{
				{
					Name:       "openshift-sdn",
					Interface:  "eth0",
					IPs:        []string{"10.130.0.16"},
					Mac:        "",
					Default:    true,
					DNS:        networkv1.DNS{Nameservers: nil, Domain: "", Search: nil, Options: nil},
					DeviceInfo: nil,
					Gateway:    nil,
				},
				{
					Name:       "openstack/internalapi",
					Interface:  "net1",
					IPs:        []string{"172.17.0.226"},
					Mac:        "a2:ef:bb:ae:65:45",
					Default:    false,
					DNS:        networkv1.DNS{Nameservers: nil, Domain: "", Search: nil, Options: nil},
					DeviceInfo: nil,
					Gateway:    nil,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			networkStatus, err := GetNetworkStatusFromAnnotation(tt.annotations)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(networkStatus).To(HaveLen(len(tt.want)))
			g.Expect(networkStatus).To(BeEquivalentTo(tt.want))
		})
	}

}

func TestGetNetworkIFName(t *testing.T) {

	tests := []struct {
		name string
		nad  string
		want string
	}{
		{
			name: "short NAD name",
			nad:  "short",
			want: "short",
		},
		{
			name: "long NAD name",
			nad:  "reallylongnadnamewithmorethan15chars",
			want: "reallylongnadna",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			ifName := GetNetworkIFName(tt.nad)

			g.Expect(ifName).To(BeEquivalentTo(tt.want))
		})
	}
}

func TestEnsureNetworksAnnotation(t *testing.T) {

	tests := []struct {
		name    string
		nadList []networkv1.NetworkAttachmentDefinition
		want    map[string]string
	}{
		{
			name:    "No network",
			nadList: []networkv1.NetworkAttachmentDefinition{},
			want:    map[string]string{networkv1.NetworkAttachmentAnnot: "[]"},
		},
		{
			name: "Single network",
			nadList: []networkv1.NetworkAttachmentDefinition{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "one", Namespace: "foo"},
					Spec: networkv1.NetworkAttachmentDefinitionSpec{
						Config: `
{
  "cniVersion": "0.3.1",
  "name": "internalapi",
  "type": "macvlan",
  "master": "internalapi",
  "ipam": {
    "type": "whereabouts",
    "range": "172.17.0.0/24",
    "range_start": "172.17.0.30",
    "range_end": "172.17.0.70"
  }
}
`,
					},
				},
			},
			want: map[string]string{networkv1.NetworkAttachmentAnnot: "[{\"name\":\"one\",\"namespace\":\"foo\",\"interface\":\"one\"}]"},
		},
		{
			name: "Multiple networks",
			nadList: []networkv1.NetworkAttachmentDefinition{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "one", Namespace: "foo"},
					Spec: networkv1.NetworkAttachmentDefinitionSpec{
						Config: `
{
  "cniVersion": "0.3.1",
  "name": "internalapi",
  "type": "macvlan",
  "master": "internalapi",
  "ipam": {
    "type": "whereabouts",
    "range": "172.17.0.0/24",
    "range_start": "172.17.0.30",
    "range_end": "172.17.0.70"
  }
}
`,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "two", Namespace: "foo"},
					Spec: networkv1.NetworkAttachmentDefinitionSpec{
						Config: `
{
  "cniVersion": "0.3.1",
  "name": "tenant",
  "type": "macvlan",
  "master": "tenant",
  "ipam": {
    "type": "whereabouts",
    "range": "172.19.0.0/24",
    "range_start": "172.19.0.30",
    "range_end": "172.19.0.70"
  }
}
`,
					},
				},
			},
			want: map[string]string{networkv1.NetworkAttachmentAnnot: "[{\"name\":\"one\",\"namespace\":\"foo\",\"interface\":\"one\"},{\"name\":\"two\",\"namespace\":\"foo\",\"interface\":\"two\"}]"},
		},
		{
			name: "With gateway defined",
			nadList: []networkv1.NetworkAttachmentDefinition{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "one", Namespace: "foo"},
					Spec: networkv1.NetworkAttachmentDefinitionSpec{
						Config: `
{
  "cniVersion": "0.3.1",
  "name": "internalapi",
  "type": "macvlan",
  "master": "internalapi",
  "ipam": {
    "type": "whereabouts",
    "range": "172.17.0.0/24",
    "range_start": "172.17.0.30",
    "range_end": "172.17.0.70",
    "gateway": "172.17.0.1"
  }
}
`,
					},
				},
			},
			want: map[string]string{networkv1.NetworkAttachmentAnnot: "[{\"name\":\"one\",\"namespace\":\"foo\",\"interface\":\"one\",\"default-route\":[\"172.17.0.1\"]}]"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			networkAnnotation, err := EnsureNetworksAnnotation(tt.nadList)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(networkAnnotation).To(HaveLen(len(tt.want)))
			g.Expect(networkAnnotation).To(BeEquivalentTo(tt.want))
		})
	}
}

func TestGetJSONPathFromConfig(t *testing.T) {

	tests := []struct {
		name string
		nad  networkv1.NetworkAttachmentDefinition
		path string
		want string
	}{
		{
			name: "No config",
			nad:  networkv1.NetworkAttachmentDefinition{},
			path: ".ipam",
			want: "",
		},
		{
			name: "get .name",
			nad: networkv1.NetworkAttachmentDefinition{
				ObjectMeta: metav1.ObjectMeta{Name: "one", Namespace: "foo"},
				Spec: networkv1.NetworkAttachmentDefinitionSpec{
					Config: `
			{
			  "cniVersion": "0.3.1",
			  "name": "internalapi",
			  "type": "macvlan",
			  "master": "internalapi",
			  "ipam": {
			    "type": "whereabouts",
			    "range": "172.17.0.0/24",
			    "range_start": "172.17.0.30",
			    "range_end": "172.17.0.70"
			  }
			}
			`,
				},
			},
			path: ".name",
			want: "internalapi",
		},
		{
			name: "get .ipam.range",
			nad: networkv1.NetworkAttachmentDefinition{
				ObjectMeta: metav1.ObjectMeta{Name: "one", Namespace: "foo"},
				Spec: networkv1.NetworkAttachmentDefinitionSpec{
					Config: `
			{
			  "cniVersion": "0.3.1",
			  "name": "internalapi",
			  "type": "macvlan",
			  "master": "internalapi",
			  "ipam": {
			    "type": "whereabouts",
			    "range": "172.17.0.0/24",
			    "range_start": "172.17.0.30",
			    "range_end": "172.17.0.70"
			  }
			}
			`,
				},
			},
			path: ".ipam.range",
			want: "172.17.0.0/24",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			networkAnnotation, err := GetJSONPathFromConfig(tt.nad, tt.path)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(networkAnnotation).To(HaveLen(len(tt.want)))
			g.Expect(networkAnnotation).To(BeEquivalentTo(tt.want))
		})
	}
}
