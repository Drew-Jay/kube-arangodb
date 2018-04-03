//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Copyright holder is ArangoDB GmbH, Cologne, Germany
//
// Author Ewout Prangsma
//

package probe

import (
	"net/http"
	"sync/atomic"
)

// Probe wraps a readiness probe handler.
type Probe struct {
	ready int32
}

// SetReady marks the probe as ready.
func (p *Probe) SetReady() {
	atomic.StoreInt32(&p.ready, 1)
}

// ReadyHandler writes back the HTTP status code 200 if the operator is ready, and 500 otherwise.
func (p *Probe) ReadyHandler(w http.ResponseWriter, r *http.Request) {
	isReady := atomic.LoadInt32(&p.ready) != 0
	if isReady {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
