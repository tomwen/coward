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

package common

import "sync"

// Counter counts number atomicly
type Counter interface {
	Add(n uint64) uint64
	Remove(n uint64) uint64
	Load(loader func(d uint64))
	LoadThenAdd(callback func(uint64), u uint64) uint64
	LoadThenRemove(callback func(uint64), u uint64) uint64
}

// counter implements Counter
type counter struct {
	data   uint64
	locker sync.RWMutex
}

// NewCounter creates a new counter
func NewCounter(defaultD uint64) Counter {
	return &counter{
		data:   defaultD,
		locker: sync.RWMutex{},
	}
}

// Add adds number to the counter
func (c *counter) Add(u uint64) uint64 {
	c.locker.Lock()

	defer c.locker.Unlock()

	c.data += u

	return c.data
}

// Remove subtract number from the counter
func (c *counter) Remove(u uint64) uint64 {
	c.locker.Lock()

	defer c.locker.Unlock()

	c.data -= u

	return c.data
}

// Load call the callback to read the counter
func (c *counter) Load(loader func(d uint64)) {
	c.locker.RLock()

	defer c.locker.RUnlock()

	loader(c.data)
}

// LoadThenAdd loads current counter result and add number to the counter
func (c *counter) LoadThenAdd(callback func(uint64), u uint64) uint64 {
	c.locker.Lock()

	defer c.locker.Unlock()

	callback(c.data)

	c.data += u

	return c.data
}

// LoadThenAdd loads current counter result and remove number from the counter
func (c *counter) LoadThenRemove(callback func(uint64), u uint64) uint64 {
	c.locker.Lock()

	defer c.locker.Unlock()

	callback(c.data)

	c.data -= u

	return c.data
}
