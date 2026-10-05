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
	"sync/atomic"
	"testing"
	"time"
)

// blockingPuller returns a Puller whose pulls block until release receives a value.
// Each started pull is reported on started with its checkLastUpdate flag.
func blockingPuller(t *testing.T) (puller *Puller, started chan bool, release chan struct{}, running *atomic.Int64) {
	started = make(chan bool, 100)
	release = make(chan struct{})
	running = &atomic.Int64{}
	puller = &Puller{}
	puller.pull = func(checkLastUpdate bool) {
		if running.Add(1) > 1 {
			t.Error("pulls overlap")
		}
		started <- checkLastUpdate
		<-release
		running.Add(-1)
	}
	return
}

func expectStarted(t *testing.T, started chan bool) bool {
	t.Helper()
	select {
	case check := <-started:
		return check
	case <-time.After(time.Second):
		t.Fatal("expected a pull to start")
		return false
	}
}

func expectNotStarted(t *testing.T, started chan bool) {
	t.Helper()
	select {
	case <-started:
		t.Fatal("unexpected pull")
	case <-time.After(100 * time.Millisecond):
	}
}

func waitOrFail(t *testing.T, wg *sync.WaitGroup) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("callers still waiting, more pulls than expected")
	}
}

func TestPullerCallersDuringARunningPullShareTheNextPull(t *testing.T) {
	puller, started, release, _ := blockingPuller(t)

	first := sync.WaitGroup{}
	first.Add(1)
	go func() {
		defer first.Done()
		puller.Pull(true)
	}()
	expectStarted(t, started)

	waiting := sync.WaitGroup{}
	returned := atomic.Int64{}
	for range 20 {
		waiting.Add(1)
		go func() {
			defer waiting.Done()
			puller.Pull(true)
			returned.Add(1)
		}()
	}
	time.Sleep(100 * time.Millisecond)
	expectNotStarted(t, started)

	release <- struct{}{}
	waitOrFail(t, &first)
	if returned.Load() != 0 {
		t.Fatal("callers returned after a pull that started before their call")
	}

	//exactly one following pull for all 20 callers
	expectStarted(t, started)
	release <- struct{}{}
	expectNotStarted(t, started)
	waitOrFail(t, &waiting)
}

func TestPullerWithoutTimestampCheckWins(t *testing.T) {
	puller, started, release, _ := blockingPuller(t)

	go puller.Pull(true)
	expectStarted(t, started)

	done := sync.WaitGroup{}
	for _, check := range []bool{true, false, true} {
		done.Add(1)
		go func() {
			defer done.Done()
			puller.Pull(check)
		}()
	}
	time.Sleep(100 * time.Millisecond)
	release <- struct{}{}

	if expectStarted(t, started) {
		t.Error("the shared pull has to skip the timestamp check if one caller asked for that")
	}
	release <- struct{}{}
	waitOrFail(t, &done)
}

func TestPullerCallersAfterAFinishedPullGetANewOne(t *testing.T) {
	puller, started, release, _ := blockingPuller(t)
	for range 3 {
		go puller.Pull(true)
		expectStarted(t, started)
		release <- struct{}{}
	}
}

func TestPullerConcurrentCallersNeverOverlap(t *testing.T) {
	running := atomic.Int64{}
	pulls := atomic.Int64{}
	puller := &Puller{}
	puller.pull = func(checkLastUpdate bool) {
		if running.Add(1) > 1 {
			t.Error("pulls overlap")
		}
		pulls.Add(1)
		time.Sleep(time.Millisecond)
		running.Add(-1)
	}
	wg := sync.WaitGroup{}
	for range 200 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			puller.Pull(true)
		}()
	}
	waitOrFail(t, &wg)
	if pulls.Load() >= 200 {
		t.Error("callers did not share pulls", pulls.Load())
	}
}
