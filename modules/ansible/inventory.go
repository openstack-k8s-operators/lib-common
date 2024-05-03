/*
Copyright 2022.

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

package ansible

import "gopkg.in/yaml.v3"

// Host represents ansible host
type Host struct {
	name string
	Vars map[string]interface{} `yaml:",inline"`
}

// MakeHost instatiates an Host object
func MakeHost(name string) Host {
	return Host{
		name: name,
		Vars: make(map[string]interface{}),
	}
}

// Group represents ansible group
type Group struct {
	name     string
	Vars     map[string]interface{} `yaml:",omitempty"`
	Hosts    map[string]Host        `yaml:",omitempty"`
	Children map[string]Group       `yaml:",omitempty"`
}

// AddHost adds a host to the current group
func (group Group) AddHost(name string) Host {
	host := MakeHost(name)
	group.Hosts[name] = host
	return host
}

// MakeGroup instatiates an Group object
func MakeGroup(name string) Group {
	return Group{
		name:     name,
		Vars:     make(map[string]interface{}),
		Hosts:    make(map[string]Host),
		Children: make(map[string]Group),
	}
}

// AddChild adds a child group to the current group
func (group Group) AddChild(newGroup Group) Group {
	group.Children[newGroup.name] = newGroup
	return newGroup
}

// Inventory contains parsed inventory representation
type Inventory struct {
	Groups map[string]Group `yaml:",inline,flow"`
}

// MakeInventory instatiates an Inventory object
func MakeInventory() Inventory {
	return Inventory{
		Groups: make(map[string]Group),
	}
}

// AddGroup adds a group to the current inventory
func (inv Inventory) AddGroup(name string) Group {
	group := MakeGroup(name)
	inv.Groups[name] = group
	return group
}

// MarshalYAML serializes the Inventory object
func (inv *Inventory) MarshalYAML() ([]byte, error) {
	return yaml.Marshal(inv)
}

// UnmarshalYAML deserializes into the Inventory object
func UnmarshalYAML(in []byte) (Inventory, error) {
	var out Inventory
	// TODO: fix Host/Group empty names
	err := yaml.Unmarshal(in, &out)
	return out, err
}
