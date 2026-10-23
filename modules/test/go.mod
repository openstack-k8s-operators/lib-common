module github.com/openstack-k8s-operators/lib-common/modules/test

go 1.26.3

require (
	github.com/go-logr/logr v1.4.4
	github.com/onsi/gomega v1.44.0
	golang.org/x/mod v0.41.0
)

require (
	github.com/google/go-cmp v0.7.0 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/net v0.60.0 // indirect
	golang.org/x/text v0.42.0 // indirect
)

replace github.com/openstack-k8s-operators/lib-common/modules/common => ../common

replace github.com/openstack-k8s-operators/lib-common/modules/openstack => ../openstack

// mschuppert: map to latest commit from release-4.22 tag
// must consistent within modules and service operators
replace github.com/openshift/api => github.com/openshift/api v0.0.0-20261007145850-7e4a91f5566b
