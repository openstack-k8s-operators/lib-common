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

import (
	"testing"
)

func TestInventoryMarshallBasic(t *testing.T) {
	inventory := MakeInventory()
	all := inventory.AddGroup("all")
	host := all.AddHost("testing")
	host.Vars["ansible_host"] = "host.test"
	childTest := all.AddChild(MakeGroup("child_test"))
	childHost := childTest.AddHost("child_testing")
	childHost.Vars["ansible_host"] = "child.host.test"

	invData, err := inventory.MarshalYAML()
	if err != nil {
		t.Log("error should be nil", err)
		t.Fail()
	}
	expected := `all:
    hosts:
        testing:
            ansible_host: host.test
    children:
        child_test:
            hosts:
                child_testing:
                    ansible_host: child.host.test
`
	result := string(invData)
	if result != expected {
		t.Log("error should be"+expected+", but got", result)
		t.Fail()
	}
}

func TestInventoryMarshallNestedChildren(t *testing.T) {
	inventory := MakeInventory()
	all := inventory.AddGroup("allovercloud")
	// host := all.AddHost("testing")
	// host.Vars["ansible_host"] = "host.test"
	childTest := all.AddChild(MakeGroup("overcloud"))
	// childHost := childTest.AddHost("child_testing")
	// childHost.Vars["ansible_host"] = "child.host.test"
	compTest := childTest.AddChild(MakeGroup("Compute"))
	compHost := compTest.AddHost("192.168.0.1")
	compHost.Vars["ansible_ssh_user"] = "root"

	invData, err := inventory.MarshalYAML()
	if err != nil {
		t.Log("error should be nil", err)
		t.Fail()
	}
	expected := `allovercloud:
    children:
        overcloud:
            children:
                Compute:
                    hosts:
                        192.168.0.1:
                            ansible_ssh_user: root
`
	result := string(invData)
	if result != expected {
		t.Log("error should be"+expected+", but got", result)
		t.Fail()
	}
}

func TestInventoryUnmarshallBasic(t *testing.T) {
	expected := MakeInventory()
	all := expected.AddGroup("all")
	host := all.AddHost("testing")
	host.Vars["ansible_host"] = "host.test"
	childTest := all.AddChild(MakeGroup("child_test"))
	childHost := childTest.AddHost("child_testing")
	childHost.Vars["ansible_host"] = "child.host.test"

	inventory := `all:
    hosts:
        testing:
            ansible_host: host.test
    children:
        child_test:
            hosts:
                child_testing:
                    ansible_host: child.host.test
`
	invObj, _ := UnmarshalYAML([]byte(inventory))

	if invObj.Groups["all"].Hosts["testing"].Vars["ansible_host"] != expected.Groups["all"].Hosts["testing"].Vars["ansible_host"] {
		t.Logf("error should be %+v but got %+v", expected.Groups["all"].Hosts, invObj.Groups["all"].Hosts)
		t.Fail()
	}
	if invObj.Groups["all"].Children["child_test"].Hosts["child_testing"].Vars["ansible_host"] != expected.Groups["all"].Children["child_test"].Hosts["child_testing"].Vars["ansible_host"] {
		t.Logf("error should be %+v but got %+v", expected.Groups["all"].Children, invObj.Groups["all"].Children)
		t.Fail()
	}

}

func TestInventoryUnmarshallNestedChildren(t *testing.T) {
	expected := MakeInventory()
	all := expected.AddGroup("allovercloud")
	host := all.AddHost("testing")
	host.Vars["ansible_host"] = "host.test"
	childTest := all.AddChild(MakeGroup("overcloud"))
	childHost := childTest.AddHost("child_testing")
	childHost.Vars["ansible_host"] = "child.host.test"
	compTest := childTest.AddChild(MakeGroup("Compute"))
	compHost := compTest.AddHost("192.168.0.1")
	compHost.Vars["ansible_ssh_user"] = "root"

	inventory := `allovercloud:
    children:
        overcloud:
            children:
                Compute:
                    hosts:
                        192.168.0.1:
                            ansible_ssh_user: root
`
	invObj, _ := UnmarshalYAML([]byte(inventory))

	hosts := invObj.Groups["allovercloud"].Children["overcloud"].Children["Compute"].Hosts
	expectedHosts := expected.Groups["allovercloud"].Children["overcloud"].Children["Compute"].Hosts

	if hosts["192.168.0.1"].Vars["ansible_ssh_user"] != expectedHosts["192.168.0.1"].Vars["ansible_ssh_user"] {
		t.Logf("error should be %+v but got %+v", expected.Groups["allovercloud"].Children, invObj.Groups["allovercloud"].Children)
		t.Fail()
	}
}
