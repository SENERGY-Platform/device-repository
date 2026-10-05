/*
 * Copyright 2026 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mgwmirror

import (
	"sync"

	"github.com/SENERGY-Platform/device-repository/v3/lib/configuration"
	"github.com/SENERGY-Platform/device-repository/v3/lib/database"
)

// Puller runs the pulls of one mirror one after another.
// Two pulls at the same time would list, write and remove the same collections concurrently;
// the removal of one could act on a listing the other has already changed.
//
// A running pull may already have passed the collection a caller is interested in,
// so Pull() waits for a pull that starts after the call. All callers arriving while a pull runs
// share the one following pull, so a hanging source does not queue up a pull per caller.
type Puller struct {
	config configuration.Config
	db     database.Database
	pull   func(checkLastUpdate bool)

	runMux sync.Mutex //held while a pull runs

	pendingMux sync.Mutex   //guards pending
	pending    *pendingPull //the next pull, not started yet; nil if nobody waits for one
}

type pendingPull struct {
	checkLastUpdate bool
	done            chan struct{}
}

func NewPuller(config configuration.Config, db database.Database) *Puller {
	result := &Puller{config: config, db: db}
	result.pull = func(checkLastUpdate bool) {
		pull(result.config, result.db, checkLastUpdate)
	}
	return result
}

// Pull returns after a pull that started after the call has finished.
func (this *Puller) Pull(checkLastUpdate bool) {
	this.pendingMux.Lock()
	p := this.pending
	if p != nil {
		//a pull without timestamp check covers one with
		p.checkLastUpdate = p.checkLastUpdate && checkLastUpdate
		this.pendingMux.Unlock()
		<-p.done
		return
	}
	p = &pendingPull{checkLastUpdate: checkLastUpdate, done: make(chan struct{})}
	this.pending = p
	this.pendingMux.Unlock()

	this.runMux.Lock()
	defer close(p.done)
	defer this.runMux.Unlock()

	this.pendingMux.Lock()
	//from here on, the pull counts as started: new callers wait for the following one
	this.pending = nil
	checkLastUpdate = p.checkLastUpdate
	this.pendingMux.Unlock()

	this.pull(checkLastUpdate)
}
