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

package apis

import (
	"fmt"
	"net/http"

	"github.com/go-logr/logr"
)

// Handler defines which URL patter is handled by which function
type Handler struct {
	// Pattern is the request URL to handle. If two patters are matching the
	// same request then the handler for the longer pattern will be executed.
	// Using the same pattern in two handlers will cause a panic.
	Pattern string
	// Func the the function that handles the request by writing a response
	Func func(http.ResponseWriter, *http.Request)
}

// APIFixture is a base struct to implement OpenStack API simulators for the
// EnvTest.
type APIFixture struct {
	Log        logr.Logger
	Server     *FakeAPIServer
	OwnsServer bool
	URLBase    string
}

// LogRequest logs details of the http.Request
func (f *APIFixture) LogRequest(r *http.Request) {
	f.Log.Info("OpenStack API request", "method", r.Method, "URI", r.RequestURI)
}

// Cleanup stops the embedded http server if it was created by the fixture
// during setup
func (f *APIFixture) Cleanup() {
	if f.OwnsServer {
		f.Server.Cleanup()
	}
}

// Endpoint is the URL the fixture's embedded http server listening on
func (f *APIFixture) Endpoint() string {
	return f.Server.Endpoint() + f.URLBase
}

// UnexpectedRequest sends a HTTP 500 response. Use it if the request is
// understood but not implemented by the fixture
func (f *APIFixture) UnexpectedRequest(w http.ResponseWriter, r *http.Request) {
	f.Log.Info("Unexpected OpenStackAPI request", "method", r.Method, "URI", r.RequestURI)
	w.WriteHeader(500)
	_, _ = fmt.Fprintf(w, "Unexpected OpenStackAPI request %s %s", r.Method, r.RequestURI)
}

// InternalError sends a HTTP 500 response Use it if there was an unexpected
// error during the request processing
func (f *APIFixture) InternalError(err error, msg string, w http.ResponseWriter, r *http.Request) {
	f.Log.Info("Internal error", "method", r.Method, "URI", r.RequestURI, "error", err, "message", msg)
	w.WriteHeader(500)
	_, _ = fmt.Fprintf(w, "Internal error in %s %s: %s: %v", r.Method, r.RequestURI, msg, err)
}
