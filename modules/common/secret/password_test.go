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
	"regexp"
	"slices"
	"testing"

	. "github.com/onsi/gomega" // nolint:revive
)

const ErrMsg string = "password does not meet the requirements"

// TestVector -
type TestVector struct {
	name     string
	password string
	wantErr  bool
	errMsg   string
}

// Pattern definition (requirements and reject rules)
var testRequirements []Rule = []Rule{
	{
		description: "Must contain at least one digit",
		pattern:     *regexp.MustCompile(`.*\d.*`),
	},
	{
		description: "Must contain at least one lowercase letter",
		pattern:     *regexp.MustCompile(`.*[a-z].*`),
	},
	{
		description: "Must contain at least one uppercase letter",
		pattern:     *regexp.MustCompile(`.*[A-Z].*`),
	},
	{
		description: "Must be at least 8 characters long",
		pattern:     *regexp.MustCompile(`^.{8,}$`),
	},
}

var testRejects []Rule = []Rule{
	{
		description: "Password contains shell expansion patterns ($variable, ${variable}, $(command), `command`)",
		pattern:     *regexp.MustCompile(`\$[A-Za-z_][A-Za-z0-9_]*|\$\{[^}]*\}|\$\([^)]*\)|` + "`[^`]*`"),
	},
}

// Input TestVectors
// 1. Fernet pattern
// 2. alphaNumPattern
// 3. validPattern
// 4. Invalid patterns
var fernetPattern []TestVector = []TestVector{
	// Fernet pattern:
	// python3 -c "from cryptography.fernet import Fernet; print(Fernet.generate_key().decode(' UTF-8'))";
	{
		name:     "Fernet Pattern Test 1",
		password: "dzUSgMefMDi3Mrq-PF8p4lHoOdF-ps_oDQB9-KzS-j0=",
		wantErr:  false,
	},
	{
		name:     "Fernet Pattern Test 2",
		password: "7Q0cfVtxqzZMKWhY5LJQF4sImYyOvC_BYYd-yg2GvUg=",
		wantErr:  false,
	},
	{
		name:     "Fernet Pattern Test 3",
		password: "UjQSpLh2WGEIFZ1Y-QCSiUr4aE76Wu3YdYLTStyEK1c=",
		wantErr:  false,
	},
	{
		name:     "Fernet Pattern Test 4",
		password: "UWtR_BCTgszn2kDhz_yxBoxxiHytMB1IR0t200uRD2s=",
		wantErr:  false,
	},
	{
		name:     "Fernet Pattern Test 5",
		password: "7cXj_fYimZ1WNu_87kNj4WM6JaI2KqCL3In2WZhAD7I=",
		wantErr:  false,
	},
}

var alphaNumPattern []TestVector = []TestVector{
	// $(tr -dc 'A-Za-z0-9' < /dev/urandom | head -c 32 | base64)
	{
		name:     "Default pattern 1",
		password: "QTFkZ1U2VnF3RWdXWjZuQW9teUo2dHlEUk43UWtaOHM=",
		wantErr:  false,
	},
	{
		name:     "Default pattern 2",
		password: "SlBVUm9OUVhEOWpjOHJsNkJmTENNbnlENjJvMU5yc1A=",
		wantErr:  false,
	},
	{
		name:     "Default pattern 3",
		password: "QzA5ejh5a1ZoWVIxbk1LUmZ4elJNMVBQNFNVdzdhaG0=",
		wantErr:  false,
	},
	{
		name:     "Default pattern 4",
		password: "ckIwWk1MUnRsSmtqdGFseThmRkYyUlhUUDdXUXhOcmc=",
		wantErr:  false,
	},
	{
		name:     "Default pattern 5",
		password: "aUxHdDRDcFR4amM2MXg4bVBoaGU2blBJRHFER3hoMFU=",
		wantErr:  false,
	},
}

var validPattern []TestVector = []TestVector{
	// Valid password scenarios
	{
		name:     "valid password with all requirements",
		password: "Password123",
		wantErr:  false,
	},
	{
		name:     "valid password with minimum length",
		password: "Abcdef12",
		wantErr:  false,
	},
	{
		name:     "valid password with longer length",
		password: "MySecurePassword123",
		wantErr:  false,
	},
	{
		name:     "valid password with allowed special characters",
		password: "Password123!@+=._-",
		wantErr:  false,
	},
	{
		name:     "valid password with semicolon",
		password: "Password123;allowed",
		wantErr:  false,
	},
	{
		name:     "valid password with angle brackets",
		password: "Password123<valid>",
		wantErr:  false,
	},
	{
		name:     "valid password with caret character",
		password: "Password123^valid",
		wantErr:  false,
	},
	{
		name:     "valid password with percent character",
		password: "Password123%valid",
		wantErr:  false,
	},
	{
		name:     "valid password with safe metacharacters",
		password: "Password123*?{}[]|&~#'\"",
		wantErr:  false,
	},
	{
		name:     "valid password with isolated dollar and number",
		password: "Password123$123",
		wantErr:  false,
	},
}

var invalidPattern []TestVector = []TestVector{
	// Invalid password scenarios - shell expansion patterns
	{
		name:     "password made by all numbers",
		password: "12345678",
		wantErr:  true,
	},
	{
		name:     "empty password",
		password: "",
		wantErr:  true,
		errMsg:   "empty password not allowed",
	},
	{
		name:     "password without uppercase",
		password: "password123",
		wantErr:  true,
	},
	{
		name:     "password without lowercase",
		password: "PASSWORD123",
		wantErr:  true,
	},
	{
		name:     "password without digit",
		password: "PasswordABC",
		wantErr:  true,
	},
	{
		name:     "password too short",
		password: "Pass12",
		wantErr:  true,
	},
	{
		name:     "password with variable expansion",
		password: "Password123$HOME",
		wantErr:  true,
		errMsg:   ErrMsg,
	},
	{
		name:     "password with variable expansion underscore",
		password: "Password123$USER_NAME",
		wantErr:  true,
		errMsg:   ErrMsg,
	},
	{
		name:     "password with braced variable expansion",
		password: "Password123${HOME}",
		wantErr:  true,
		errMsg:   ErrMsg,
	},
	{
		name:     "password with command substitution",
		password: "Password123$(echo bad)",
		wantErr:  true,
		errMsg:   ErrMsg,
	},
	{
		name:     "password with backtick command substitution",
		password: "Password123`echo bad`",
		wantErr:  true,
		errMsg:   ErrMsg,
	},
	{
		name:     "password with complex variable expansion",
		password: "Password123${VARIABLE_NAME}",
		wantErr:  true,
		errMsg:   ErrMsg,
	},
	{
		name:     "shell expansion attack example",
		password: "c^sometext02%text%text02$someText&",
		wantErr:  true,
		errMsg:   ErrMsg,
	},
}

func TestValidatePassword(t *testing.T) {
	// Save original values
	originalRequirements := requirements
	originalRejects := rejects

	// Validate testRequirements and testRejects rules
	requirements = testRequirements
	rejects = testRejects

	// Restore after test
	defer func() {
		requirements = originalRequirements
		rejects = originalRejects
	}()

	// Build TestVector
	var tests []TestVector
	tests = slices.Concat(
		tests,
		validPattern,
		invalidPattern,
		alphaNumPattern,
		fernetPattern,
	)

	validator := PasswordValidator{}

	// Execute ValidatePassword against the generated TestVector
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			err := validator.Validate(tt.password)

			if tt.wantErr {
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(ContainSubstring(tt.errMsg))
			} else {
				g.Expect(err).ToNot(HaveOccurred())
			}
		})
	}
}

func TestValidatePasswordWithCustomRules(t *testing.T) {
	// Define custom requirements (strict rules)
	customRequirements := []Rule{
		{
			description: "Must only contain alphanumerical and safe special characters",
			pattern:     *regexp.MustCompile(`^[a-zA-Z0-9@#%^*-_=+:,.!~]+$`),
		},
	}

	// Define custom rejects (forbid specific patterns)
	customRejects := []Rule{
		{
			description: "Must not contain carriage return",
			pattern:     *regexp.MustCompile(`[\n]`),
		},
	}

	tests := []struct {
		name      string
		validator PasswordValidator
		password  string
		wantErr   bool
		errMsg    string
	}{
		// Test with nil rules (should use defaults)
		{
			name:      "nil rules uses defaults - valid",
			validator: PasswordValidator{},
			password:  "ValidPassword123",
			wantErr:   false,
		},
		{
			name:      "nil rules uses defaults - shell expansion rejected",
			validator: PasswordValidator{},
			password:  "Password123$HOME",
			wantErr:   true,
			errMsg:    ErrMsg,
		},
		{
			name:      "nil rules uses defaults - empty password rejected",
			validator: PasswordValidator{},
			password:  "",
			wantErr:   true,
			errMsg:    "empty password not allowed",
		},

		// Test with custom requirements only
		{
			name: "custom requirements - safe special characters",
			validator: PasswordValidator{
				Requirements: &customRequirements,
			},
			password: "#S3cure!Pass#",
			wantErr:  false,
		},

		// Test with custom rejects only
		{
			name: "custom rejects - must not contains '\n'",
			validator: PasswordValidator{
				Rejects: &customRejects,
			},
			password: "MyPASSWORD123\n",
			wantErr:  true,
			errMsg:   ErrMsg,
		},

		// Test with both custom requirements and rejects
		{
			name: "custom both - fully valid password",
			validator: PasswordValidator{
				Requirements: &customRequirements,
				Rejects:      &customRejects,
			},
			password: "MyS3cure!Pass",
			wantErr:  false,
		},

		// Test empty requirements/rejects slices
		{
			name: "empty requirements slice - allows any non-empty password",
			validator: PasswordValidator{
				Requirements: &[]Rule{},
			},
			password: "a",
			wantErr:  false,
		},
		{
			name: "empty rejects slice - no rejections",
			validator: PasswordValidator{
				Rejects: &[]Rule{},
			},
			password: "anything$HOME$(cmd)`test`",
			wantErr:  false,
		},
		{
			name: "both empty - allows any non-empty password",
			validator: PasswordValidator{
				Requirements: &[]Rule{},
				Rejects:      &[]Rule{},
			},
			password: "x",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			err := tt.validator.Validate(tt.password)

			if tt.wantErr {
				g.Expect(err).To(HaveOccurred())
				if tt.errMsg != "" {
					g.Expect(err.Error()).To(ContainSubstring(tt.errMsg))
				}
			} else {
				g.Expect(err).ToNot(HaveOccurred())
			}
		})
	}
}
