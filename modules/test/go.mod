module github.com/openstack-k8s-operators/lib-common/modules/test

go 1.24.4

require (
	github.com/go-logr/logr v1.4.3
	github.com/onsi/gomega v1.38.2
	golang.org/x/mod v0.27.0
)

require (
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/onsi/ginkgo/v2 v2.26.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/net v0.43.0 // indirect
	golang.org/x/text v0.28.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)

replace github.com/openstack-k8s-operators/lib-common/modules/common => ../common

replace github.com/openstack-k8s-operators/lib-common/modules/openstack => ../openstack

// mschuppert: map to latest commit from release-4.18 tag
// must consistent within modules and service operators
replace github.com/openshift/api => github.com/openshift/api v0.0.0-20250711200046-c86d80652a9e
