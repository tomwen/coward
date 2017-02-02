//  Crypto-Obscured Forwarder
//
//  Copyright (C) 2017 NI Rui <nickriose@gmail.com>
//
//  This file is part of Crypto-Obscured Forwarder.
//
//  Crypto-Obscured Forwarder is free software: you can redistribute it
//  and/or modify it under the terms of the GNU General Public License
//  as published by the Free Software Foundation, either version 3 of
//  the License, or (at your option) any later version.
//
//  Crypto-Obscured Forwarder is distributed in the hope that it will be
//  useful, but WITHOUT ANY WARRANTY; without even the implied warranty
//  of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU General Public License for more details.
//
//  You should have received a copy of the GNU General Public License
//  along with Crypto-Obscured Forwarder. If not, see
//  <http://www.gnu.org/licenses/>.

package locked

import (
	"sync"
	"time"
)

// TimeDuration is lockable time duration
type TimeDuration interface {
	Set(newValue time.Duration)
	Get() time.Duration
	Load(func(current time.Duration))
}

type timeDuration struct {
	value time.Duration
	lock  sync.RWMutex
}

// NewTimeDuration creates a new lockable TimeDuration
func NewTimeDuration(defaultV time.Duration) TimeDuration {
	return &timeDuration{
		value: defaultV,
		lock:  sync.RWMutex{},
	}
}

// Set sets a new value
func (t *timeDuration) Set(newValue time.Duration) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.value = newValue
}

// Get gets current value
func (t *timeDuration) Get() time.Duration {
	t.lock.RLock()
	defer t.lock.RUnlock()

	return t.value
}

// Load loads current value to a callback, and keep blocking while
// the callback is running
func (t *timeDuration) Load(callback func(current time.Duration)) {
	t.lock.RLock()
	defer t.lock.RUnlock()

	callback(t.value)
}
