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
	"net/http"
	"net/http/httptest"

	"github.com/go-logr/logr"
)

// FakeAPIServer represents an embedded http server to simulate http services
type FakeAPIServer struct {
	mux    *http.ServeMux
	server *httptest.Server
	log    logr.Logger
}

// Endpoint is the URL the embedded http server listening on
func (s *FakeAPIServer) Endpoint() string {
	return s.server.URL
}

// Setup creates and starts the embedded http server on localhost and on
// a random port
func (s *FakeAPIServer) Setup(log logr.Logger) {
	s.log = log
	s.mux = http.NewServeMux()
	s.server = httptest.NewServer(s.mux)
	// The / URL matches to every request if no handle registered with a more
	// specific URL pattern
	s.mux.HandleFunc("/", s.fallbackHandler)
}

func (s *FakeAPIServer) fallbackHandler(w http.ResponseWriter, r *http.Request) {
	if r == nil {
		s.log.Info("Unexpected OpenStackAPI nil request")
		w.WriteHeader(500)
		return
	}

	s.log.Info("Unexpected OpenStackAPI request", "method", r.Method, "URI", r.RequestURI)
	w.WriteHeader(500)
}

// Cleanup stops the embedded http server
func (s *FakeAPIServer) Cleanup() {
	s.server.Close()
}
