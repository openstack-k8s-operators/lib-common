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
	"fmt"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// ValidateStorageRequest - validates a storage request meets a provided min size. Depending on the
// err (bool) the check result is either a warning or an invalid error.
//
// example usage:
//
//   - return a warning
//
//     ValidateStorageRequest(<path>, "500M", "5G", false)
//
//   - return an error
//
//     ValidateStorageRequest(<path>, "500M", "5G", true)
func ValidateStorageRequest(basePath *field.Path, req string, minReq string, err bool) (admission.Warnings, field.ErrorList) {
	allErrs := field.ErrorList{}
	allWarn := []string{}

	// If the StorageRequest is a wrong string, we must return
	// an error. MustParse can't be used in this context as it
	// generates panic() and we can't recover the operator.
	storageRequest, parseError := resource.ParseQuantity(req)
	if parseError != nil {
		parseQuantityError := fmt.Sprintf("Field %s: %s is invalid",
			basePath.Child("storageRequest").String(), req)
		// Return error if err == true was provided, else a warning
		if err {
			allErrs = append(allErrs, field.Invalid(basePath, req, parseQuantityError))
		} else {
			allWarn = append(allWarn, parseQuantityError)
		}
		return allWarn, allErrs
	}

	storageRequestProd, parseError := resource.ParseQuantity(minReq)
	if parseError != nil {
		parseQuantityError := fmt.Sprintf("Invalid %s quantity", minReq)
		// Return error if err == true was provided, else a warning
		if err {
			allErrs = append(allErrs, field.Invalid(basePath, minReq, parseQuantityError))
		} else {
			allWarn = append(allWarn, parseQuantityError)
		}
		return allWarn, allErrs
	}

	if storageRequest.Cmp(storageRequestProd) < 0 {
		res := fmt.Sprintf("%s: %s is not appropriate for production! For production use at least %s!",
			basePath.Child("storageRequest").String(), req, minReq)
		// Return error if err == true was provided, else a warning
		if err {
			allErrs = append(allErrs, field.Invalid(basePath, req, res))
		} else {
			allWarn = append(allWarn, res)
		}
	}
	return allWarn, allErrs
}
