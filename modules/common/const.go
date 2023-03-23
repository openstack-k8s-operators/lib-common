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

package common

// consts used by all operators
const (
	// AppSelector - used by operators to specify labels
	AppSelector = "service"
	// ComponentSelector - used by operators to specify labels for a sub component
	ComponentSelector = "component"
	// CustomServiceConfigFileName - file name used to add the service customizations
	CustomServiceConfigFileName = "custom.conf"
	// DebugCommand - Default debug command for pods
	DebugCommand = "/usr/local/bin/kolla_set_configs && /bin/sleep infinity"
	// InputHashName -Name of the hash of hashes of all resources used to indentify an input change
	InputHashName = "input"
)
