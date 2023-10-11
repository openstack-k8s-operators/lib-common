/*
Copyright 2023.

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

package functional

import (
	"k8s.io/apimachinery/pkg/types"
)

type Names struct {
	Namespace            string
	IssuerName           types.NamespacedName
	SelfSignedIssuerName types.NamespacedName
	CAName               types.NamespacedName
	CertName             types.NamespacedName
}

func CreateNames(namespace string) Names {
	return Names{
		Namespace:            namespace,
		SelfSignedIssuerName: types.NamespacedName{Namespace: namespace, Name: "selfsigned"},
		CAName:               types.NamespacedName{Namespace: namespace, Name: "ca"},
		IssuerName:           types.NamespacedName{Namespace: namespace, Name: "issuer"},
		CertName:             types.NamespacedName{Namespace: namespace, Name: "cert"},
	}
}
