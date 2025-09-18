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

// Package util provides environment variable utilities
package util // nolint:revive

import "os"

// GetEnvVar - Get the value associated with key from environment variables, but use baseDefault as a value in case the ENV variable is not defined.
func GetEnvVar(key string, baseDefault string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return baseDefault
}
