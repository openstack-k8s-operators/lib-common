/*
Copyright 2025 Red Hat

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

package util

import (
	"errors"
	"fmt"
	"strings"
)

// IniOption -
type IniOption struct {
	Section string
	Key     string
	Value   string
	Unique  bool
}

// Define static errors
var (
	ErrKeyAlreadyExists     = errors.New("key already exists in section")
	ErrCouldNotPatchSection = errors.New("could not patch target section")
)

// repr - print key: value in .ini format
func (i *IniOption) repr() string {
	return fmt.Sprintf("%s = %s", i.Key, i.Value)
}

// ExtendCustomServiceConfig - customServiceConfig is tokenized and parsed in a
// loop where we keep track of two indexes:
//   - index: keep track of the current extracted token
//   - sectionIndex: when we detect a [<section>] within a token, we save the index
//     and we update it with the next section when is detected. This way we can
//     make sure to evaluate only the keys of the target section
//
// when an invalid case is detected, we return the customServiceConfig string
// unchanged, otherwise the new key=value is appended as per the IniOption struct
func ExtendCustomServiceConfig(
	iniString string,
	customServiceConfigExtend IniOption,
) (string, error) {
	// customServiceConfig is empty
	if len(iniString) == 0 {
		return iniString, nil
	}
	// Position where insert new option (-1 = target section not found)
	index := -1
	// Current section header position (-1 = no section found)
	sectionIndex := -1
	svcConfigLines := strings.Split(iniString, "\n")
	sectionName := ""
	for idx, rawLine := range svcConfigLines {
		line := strings.TrimSpace(rawLine)
		token := strings.TrimSpace(strings.SplitN(line, "=", 2)[0])

		if token == "" || strings.HasPrefix(token, "#") {
			// Skip blank lines and comments
			continue
		}
		if strings.HasPrefix(token, "[") && strings.HasSuffix(token, "]") {
			// Note the section name before looking for a backend_name
			sectionName = strings.Trim(token, "[]")
			sectionIndex = idx
			// increment the index (as an offset) only when a section is found
			if sectionName == customServiceConfigExtend.Section {
				index = idx + 1
			}
		}
		// Check if key already exists in target section
		if customServiceConfigExtend.Unique && token == customServiceConfigExtend.Key && sectionIndex > -1 &&
			sectionName == customServiceConfigExtend.Section {
			errMsg := fmt.Errorf("%w: key %s in section %s", ErrKeyAlreadyExists, token, sectionName)
			return iniString, errMsg
		}
	}
	// index didn't progress during the customServiceConfig scan:
	// return unchanged, but no error
	if index == -1 {
		errMsg := fmt.Errorf("%w: %s", ErrCouldNotPatchSection, customServiceConfigExtend.Section)
		return iniString, errMsg
	}
	// index has a valid value and it is used as a pivot to inject the new ini
	// option right after the section
	var svcExtended []string
	svcExtended = append(svcExtended, svcConfigLines[:index]...)
	svcExtended = append(svcExtended, []string{customServiceConfigExtend.repr()}...)
	svcExtended = append(svcExtended, svcConfigLines[index:]...)
	return strings.Join(svcExtended, "\n"), nil
}
