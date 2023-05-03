/*
Copyright 2023 Red Hat

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
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
)

func TestGetEnvVar(t *testing.T) {

	envVarName := fmt.Sprintf("%sTEST", t.Name())
	envVarVal := "testing"
	envVarValDefault := "default"

	t.Setenv(envVarName, envVarVal)

	tests := []struct {
		name string
		data []string
		want string
	}{
		{
			name: "Get env var where it is actually present and there is a default",
			data: []string{envVarName, envVarValDefault},
			want: envVarVal,
		},
		{
			name: "Get env var where it is actually present and there is no default",
			data: []string{envVarName, ""},
			want: envVarVal,
		},
		{
			name: "Get env var where it is not actually present",
			data: []string{"some_absent_env_var", envVarValDefault},
			want: envVarValDefault,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			val := GetEnvVar(tt.data[0], tt.data[1])

			g.Expect(val).To(BeIdenticalTo(tt.want))
		})
	}
}
