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

package log

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func defaultParams(object metav1.Object) []interface{} {
	return []interface{}{
		"ObjectType", fmt.Sprintf("%T", object),
		"ObjectNamespace", object.GetNamespace(),
		"ObjectName", object.GetName()}
}

// Info logs a non-error message to the logger with the default object
// parameters ObjectType, ObjectNamespace, ObjectName
func Info(
	logger logr.Logger,
	msg string,
	object metav1.Object,
	params ...interface{},
) {
	params = append(params, defaultParams(object)...)
	logger.Info(msg, params...)
}

// Error logs an error message to the logger with the default object
// parameters ObjectType, ObjectNamespace, ObjectName
func Error(
	logger logr.Logger,
	err error,
	msg string,
	object metav1.Object,
	params ...interface{},
) {
	params = append(params, defaultParams(object)...)
	logger.Error(err, msg, params...)
}

// IntoContext returns a new context with a logger that always logs the default
// object parameters ObjectType, ObjectNamespace, ObjectName
//
// The intended use is to upgrade the context at the start of the
// Reconcile call and use the new ctx to pass arond and use the embeded
// logger:
//
//    func (r *FooReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
//        ctx = log.IntoContext(ctx, req, "Foo")
//        // ...
//        log.FromContext(ctx).Info("Succesfully reconciled")
//        return ctrl.Result{}, nil
//    }
//
func IntoContext(ctx context.Context, req ctrl.Request, kind string) context.Context {
	return log.IntoContext(
		ctx,
		log.FromContext(
			ctx,
			"ObjectType", kind,
			"ObjectNamespace", req.Namespace,
			"ObjectName", req.Name,
		),
	)
}

// FromContext returns the logger stored in the context
func FromContext(ctx context.Context) logr.Logger {
	return log.FromContext(ctx)
}
