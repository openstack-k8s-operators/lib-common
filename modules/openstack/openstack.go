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

package openstack

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	gophercloud "github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	endpoint "github.com/openstack-k8s-operators/lib-common/modules/common/endpoint"
)

const (
	// defaultRequestTimeout is the default timeout duration for requests
	defaultRequestTimeout = 10 * time.Second
)

// OpenStack -
type OpenStack struct {
	osclient *gophercloud.ServiceClient
	region   string
	authURL  string
}

// AuthOpts -
type AuthOpts struct {
	AuthURL    string
	Username   string
	Password   string
	TenantName string
	DomainName string
	Region     string
	Scope      *gophercloud.AuthScope
	TLS        *TLSConfig
}

// TLSConfig - settings
type TLSConfig struct {
	CACerts    []string
	Insecure   bool
	ClientCert string
	ClientKey  string
}

// NewOpenStack creates a new new instance of the openstack struct from a config struct
func NewOpenStack(
	log logr.Logger,
	cfg AuthOpts,
) (*OpenStack, error) {

	opts := gophercloud.AuthOptions{
		IdentityEndpoint: cfg.AuthURL,
		Username:         cfg.Username,
		Password:         cfg.Password,
		TenantName:       cfg.TenantName,
		DomainName:       cfg.DomainName,
	}
	if cfg.Scope != nil {
		opts.Scope = cfg.Scope
	}

	// define http client for setting timeout, proxy and tls settings
	httpClient := http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
		Timeout: defaultRequestTimeout,
	}

	// create tls config
	tlsConfig := &tls.Config{}
	if cfg.TLS != nil {
		if len(cfg.TLS.CACerts) > 0 {
			caCertPool := x509.NewCertPool()
			for _, caCert := range cfg.TLS.CACerts {
				caCertPool.AppendCertsFromPEM([]byte(caCert))
			}
			tlsConfig.RootCAs = caCertPool
		}
		if cfg.TLS.Insecure {
			tlsConfig.InsecureSkipVerify = true
		}

		if cfg.TLS.ClientCert != "" && cfg.TLS.ClientKey != "" {
			cert, err := tls.LoadX509KeyPair(cfg.TLS.ClientCert, cfg.TLS.ClientKey)
			if err != nil {
				return nil, err
			}

			tlsConfig.Certificates = []tls.Certificate{cert}
		}
	}

	transport := &http.Transport{Proxy: http.ProxyFromEnvironment, TLSClientConfig: tlsConfig}

	// create provider client and add inject customized http client
	providerClient, err := openstack.NewClient(opts.IdentityEndpoint)
	if err != nil {
		return nil, err
	}

	providerClient.HTTPClient = httpClient
	providerClient.HTTPClient.Transport = transport

	// authenticate the client
	err = openstack.Authenticate(providerClient, opts)
	if err != nil {
		return nil, err
	}

	// create the identity client using previous providerClient
	endpointOpts := gophercloud.EndpointOpts{Type: "identity", Region: cfg.Region}
	identityClient, err := openstack.NewIdentityV3(providerClient, endpointOpts)
	if err != nil {
		return nil, err
	}

	os := OpenStack{
		osclient: identityClient,
		region:   cfg.Region,
		authURL:  cfg.AuthURL,
	}

	return &os, nil
}

// GetRegion - returns the region
func (o *OpenStack) GetRegion() string {
	return o.region
}

// GetAuthURL - returns the region
func (o *OpenStack) GetAuthURL() string {
	return o.authURL
}

// GetOSClient - returns the client
func (o *OpenStack) GetOSClient() *gophercloud.ServiceClient {
	return o.osclient
}

// GetAvailability - returns mapping of enpointtype to gophercloud.Availability
func GetAvailability(
	endpointInterface string,
) (gophercloud.Availability, error) {
	var availability gophercloud.Availability
	if endpointInterface == string(endpoint.EndpointAdmin) {
		availability = gophercloud.AvailabilityAdmin
	} else if endpointInterface == string(endpoint.EndpointInternal) {
		availability = gophercloud.AvailabilityInternal
	} else if endpointInterface == string(endpoint.EndpointPublic) {
		availability = gophercloud.AvailabilityPublic
	} else {
		return availability, fmt.Errorf("endpoint interface %s not known", endpointInterface)
	}
	return availability, nil
}
