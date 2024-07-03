/*
Copyright 2024 Red Hat

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

package webhook

import (
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ValidateDNS1123Label - validates a list of strings are RFC 1123 label conform. Using the
// correction parameter the validation.DNS1123LabelMaxLength (63) get reduced by the correction
// value
//
// example usage:
//
//	ValidateDNS1123Label(<path>, {"foo", "bar"}, 5)
func ValidateDNS1123Label(basePath *field.Path, keys []string, correction int) field.ErrorList {
	allErrs := field.ErrorList{}

	for _, key := range keys {
		msgs := validation.IsDNS1123Label(key)

		maxLength := validation.DNS1123LabelMaxLength - correction

		if correction > 0 && len(key) > maxLength {
			msgs = append(msgs, validation.MaxLenError(maxLength))
		}

		for _, msg := range msgs {
			allErrs = append(allErrs, field.Invalid(basePath.Key(key), key, msg))
		}
	}

	return allErrs
}
