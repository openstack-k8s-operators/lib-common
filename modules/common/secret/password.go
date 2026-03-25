/*
Copyright 2026 Red Hat

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

package secret

import (
	"errors"
	"regexp"
)

// PasswordValidator implements the Validator interface
// +kubebuilder:object:generate=false
type PasswordValidator struct {
	// Patterns that the password should adhere to
	Requirements *[]Rule
	// Patterns that are forbidden for the password
	Rejects *[]Rule
}

// Rule - pattern to match when rejecting or accepting a string
// +kubebuilder:object:generate=false
type Rule struct {
	description string
	pattern     regexp.Regexp
}

// the requirements variable defines password complexity rules that must be
// satisfied. It is currently empty to allow any password content that does
// not contain shell expansion patterns.
// Add rules as needed based on security requirements.
//
// Example requirements:
//
//	{
//	    description: "Must contain at least one digit",
//	    pattern: *regexp.MustCompile(`.*\d.*`),
//	},
//	{
//	    description: "Must contain at least one lowercase letter",
//	    pattern: *regexp.MustCompile(`.*[a-z].*`),
//	},
//	{
//	    description: "Must contain at least one uppercase letter",
//	    pattern: *regexp.MustCompile(`.*[A-Z].*`),
//	},
//	{
//	    description: "Must be at least 8 characters long",
//	    pattern: *regexp.MustCompile(`^.{8,}$`),
//	},
var requirements []Rule = []Rule{}

var rejects []Rule = []Rule{
	{
		description: "Password contains shell expansion patterns ($variable, ${variable}, $(command), `command`)",
		pattern:     *regexp.MustCompile(`\$[A-Za-z_][A-Za-z0-9_]*|\$\{[^}]*\}|\$\([^)]*\)|` + "`[^`]*`"),
	},
}

var (
	// ErrEmptyPassword -
	ErrEmptyPassword = errors.New("empty password not allowed")
	// ErrPasswordRequirements -
	ErrPasswordRequirements = errors.New("password does not meet the requirements")
)

// Validate - implements the Validator interface
// If requirements or rejects rules are not specified in the
// structure, the function uses the default rule set defined
// in this package.
func (v PasswordValidator) Validate(value string) error {
	// Check if password is empty
	if value == "" {
		return ErrEmptyPassword
	}

	reqs := &requirements
	if v.Requirements != nil {
		reqs = v.Requirements
	}
	for _, rule := range *reqs {
		if !rule.pattern.MatchString(value) {
			return ErrPasswordRequirements
		}
	}

	rejs := &rejects
	if v.Rejects != nil {
		rejs = v.Rejects
	}
	for _, rule := range *rejs {
		if rule.pattern.MatchString(value) {
			return ErrPasswordRequirements
		}
	}
	return nil
}
