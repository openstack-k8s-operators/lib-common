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

package util

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/util/rand"

	env "github.com/openstack-k8s-operators/lib-common/modules/common/env"
	corev1 "k8s.io/api/core/v1"
)

// Hash - struct to add hashes to status
type Hash struct {
	// Name of hash referencing the parameter
	Name string `json:"name,omitempty"`
	// Hash
	Hash string `json:"hash,omitempty"`
}

// ObjectHash creates a deep object hash and return it as a safe encoded string
func ObjectHash(i interface{}) (string, error) {
	// Convert the hashSource to a byte slice so that it can be hashed
	hashBytes, err := json.Marshal(i)
	if err != nil {
		return "", fmt.Errorf("unable to convert to JSON: %w", err)
	}
	hash := sha256.Sum256(hashBytes)
	return rand.SafeEncodeString(fmt.Sprint(hash)), nil
}

// SetHash - set hashStr of type hashType on hashMap if it does not exist or
// hashStr is different from current stored value. Returns hashMap and bool
// which indicates if hashMap changed.
func SetHash(
	hashMap map[string]string,
	hashType string,
	hashStr string,
) (map[string]string, bool) {
	if hashMap == nil {
		hashMap = map[string]string{}
	}
	if hash, ok := hashMap[hashType]; !ok || hash != hashStr {
		hashMap[hashType] = hashStr
		return hashMap, true
	}

	return hashMap, false
}

// HashOfInputHashes - calculates the overall hash of hashes
func HashOfInputHashes(
	hashes map[string]env.Setter,
) (string, error) {
	mergedMapVars := env.MergeEnvs([]corev1.EnvVar{}, hashes)
	hash, err := ObjectHash(mergedMapVars)
	if err != nil {
		return hash, err
	}
	return hash, nil
}
