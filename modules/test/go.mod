module github.com/openstack-k8s-operators/lib-common/modules/test

go 1.19

require (
	github.com/go-logr/logr v1.4.1
	github.com/onsi/gomega v1.30.0
	golang.org/x/mod v0.14.0
)

require (
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/pprof v0.0.0-20210720184732-4bb14d4b1be1 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/onsi/ginkgo/v2 v2.14.0 // indirect
	github.com/rogpeppe/go-internal v1.10.0 // indirect
	golang.org/x/net v0.20.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/tools v0.17.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/openstack-k8s-operators/lib-common/modules/common => ../common

replace github.com/openstack-k8s-operators/lib-common/modules/openstack => ../openstack

// mschuppert: map to latest commit from release-4.13 tag
// must consistent within modules and service operators
replace github.com/openshift/api => github.com/openshift/api v0.0.0-20230414143018-3367bc7e6ac7
