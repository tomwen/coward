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

import "sync"

// Boolean is lockable bool
type Boolean interface {
	Set(newValue bool)
	Get() bool
	GetSet(newValue bool) bool
	Load(func(current bool))
}

type boolean struct {
	value bool
	lock  sync.RWMutex
}

// NewBool creates a new lockable Boolean
func NewBool(defaultV bool) Boolean {
	return &boolean{
		value: defaultV,
		lock:  sync.RWMutex{},
	}
}

// Set sets a new value
func (b *boolean) Set(newValue bool) {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.value = newValue
}

// Get gets current value
func (b *boolean) Get() bool {
	b.lock.RLock()
	defer b.lock.RUnlock()

	return b.value
}

// Get current value of the data, and update it with a new one
func (b *boolean) GetSet(newValue bool) bool {
	b.lock.Lock()
	defer b.lock.Unlock()

	current := b.value

	b.value = newValue

	return current
}

// Load loads current value to a callback, and keep blocking while
// the callback is running
func (b *boolean) Load(callback func(current bool)) {
	b.lock.RLock()
	defer b.lock.RUnlock()

	callback(b.value)
}
