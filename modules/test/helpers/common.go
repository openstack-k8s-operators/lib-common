/*
Copyright 2022 Red Hat
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

package helpers

import (
	"context"
	"os"
	"strconv"
	"time"

	ginkgo "github.com/onsi/ginkgo/v2"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// TestHelper -
type TestHelper struct {
	k8sClient client.Client
	ctx       context.Context
	timeout   time.Duration
	interval  time.Duration
}

// NewTestHelper returns a TestHelper
func NewTestHelper(
	ctx context.Context,
	k8sClient client.Client,
	timeout time.Duration,
	interval time.Duration,
) *TestHelper {
	return &TestHelper{
		ctx:       ctx,
		k8sClient: k8sClient,
		timeout:   timeout,
		interval:  interval,
	}
}

// SkipInExistingCluster -
func SkipInExistingCluster(message string) {
	s := os.Getenv("USE_EXISTING_CLUSTER")
	v, err := strconv.ParseBool(s)

	if err == nil && v {
		ginkgo.Skip("Skipped running against existing cluster. " + message)
	}

}
