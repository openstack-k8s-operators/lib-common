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
	. "github.com/onsi/gomega" // nolint:revive
	"regexp"
	"testing"
)

const ErrMsg string = "password does not meet the requirements"

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

	tests := []struct {
		name     string
		password string
		wantErr  bool
		errMsg   string
	}{
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			err := ValidatePassword(tt.password)

			if tt.wantErr {
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(ContainSubstring(tt.errMsg))
			} else {
				g.Expect(err).ToNot(HaveOccurred())
			}
		})
	}
}
