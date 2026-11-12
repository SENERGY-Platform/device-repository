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

package docker

import (
	"testing"
	"time"
)

// WaitFor polls cond every 100 ms until it returns nil. If it does not within timeout,
// the test fails with the last error of cond.
// It replaces fixed sleeps in tests that wait for an asynchronous event (sync loop, kafka message, mirror pull).
func WaitFor(t testing.TB, timeout time.Duration, cond func() error) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for {
		err := cond()
		if err == nil {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("condition not met after %s: %v", timeout, err)
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// WaitForNextSecond returns as soon as the unix second of the system clock has changed.
// The sync jobs of the mongo collections carry a timestamp in unix seconds and are fetched only if it is
// smaller than the cutoff in unix seconds (lib/database/mongo/sync.go). Documents written in the current second
// are not fetched by a sync run that starts in the same second and wait for the next run after sync_interval.
func WaitForNextSecond() {
	now := time.Now()
	time.Sleep(now.Truncate(time.Second).Add(time.Second + 10*time.Millisecond).Sub(now))
}
