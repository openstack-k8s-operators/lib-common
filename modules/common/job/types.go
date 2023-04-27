/*
Copyright 2021 Red Hat

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

package job

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	defaultTTL int32 = 10 * 60 // 10 minutes
)

// Job -
type Job struct {
	job        *batchv1.Job
	jobType    string
	preserve   bool
	timeout    time.Duration
	beforeHash string
	hash       string
	changed    bool
}

// Helper represents the external dependencies of the Job operations
type Helper interface {
	GetBeforeObject() client.Object
	GetScheme() *runtime.Scheme
	GetLogger() logr.Logger

	CreateOrPatch(ctx context.Context, obj client.Object, f controllerutil.MutateFn) (controllerutil.OperationResult, error)
	SetControllerReference(owner, controlled metav1.Object, scheme *runtime.Scheme) error

	Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error
}
