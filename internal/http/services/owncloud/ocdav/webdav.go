// Copyright 2018-2019 CERN
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// In applying this license, CERN does not waive the privileges and immunities
// granted to it by virtue of its status as an Intergovernmental Organization
// or submit itself to any jurisdiction.

package ocdav

import (
	"context"
	"net/http"
	"path"
)

// WebDavHandler routes to the legacy dav endpoint
type WebDavHandler struct {
}

func (h *WebDavHandler) init(c *Config) error {
	return nil
}

// Handler handles requests
func (h *WebDavHandler) Handler(s *svc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			// no need for the user, and we need to be able
			// to answer preflight checks, which have no auth headers
			s.doOptions(w, r)
			return
		}

		// webdav should be death: baseURI is encoded as part of the
		// response payload in href field
		baseURI := path.Join("/", s.Prefix(), "remote.php/webdav")
		ctx := context.WithValue(r.Context(), ctxKeyBaseURI, baseURI)
		r = r.WithContext(ctx)

		switch r.Method {
		case "PROPFIND":
			s.doPropfind(w, r)
		case http.MethodHead:
			s.doHead(w, r)
		case http.MethodGet:
			s.doGet(w, r)
		case "LOCK":
			s.doLock(w, r)
		case "UNLOCK":
			s.doUnlock(w, r)
		case "PROPPATCH":
			s.doProppatch(w, r)
		case "MKCOL":
			s.doMkcol(w, r)
		case "MOVE":
			s.doMove(w, r)
		case "COPY":
			s.doCopy(w, r)
		case http.MethodPut:
			s.doPut(w, r)
		case http.MethodDelete:
			s.doDelete(w, r)
		case "REPORT":
			s.doReport(w, r)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
}
