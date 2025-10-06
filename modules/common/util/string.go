/*
Copyright 2020 Red Hat

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

package util // nolint:revive

import (
	"crypto/rand"
	"fmt"
	"time"
)

// StringInSlice - is string in slice
func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// RandomString generates a random string of specified length.
// It uses alphanumeric characters (0-9, a-z, A-Z).
func RandomString(length int) string {
	if length <= 0 {
		return ""
	}

	const charset = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	csLen := len(charset)

	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		// Fallback to a deterministic but unique string based on time
		return fmt.Sprintf("fallback-%d", time.Now().UnixNano())
	}

	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = charset[int(b[i])%csLen]
	}
	return string(result)
}
