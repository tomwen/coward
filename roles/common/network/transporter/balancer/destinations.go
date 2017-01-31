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

package balancer

import (
	"sync"

	"github.com/nickrio/coward/common"
	"github.com/nickrio/coward/roles/common/network/transporter/clients"
)

// destinations remember the best transporter client to a
// specified destination
type destinations struct {
	max          uint
	length       uint
	pole         destinationPole
	transports   clients.Clients
	destinations map[string]*destination
	destLock     sync.Mutex
}

// build creates new transports
func (d *destinations) build() *transports {
	t := transports{
		pole: transportPole{
			Head: nil,
			Tail: nil,
			Lock: sync.RWMutex{},
		},
		transports: make([]*transport, d.transports.Length()),
		sorted:     make([]*transport, d.transports.Length()),
		sortLock:   sync.RWMutex{},
		requesting: common.NewCounter(0),
	}

	tIndex := 0

	d.transports.Export(func(c clients.Client) {
		t.Register(tIndex, c)

		tIndex++
	})

	return &t
}

// expire checks if we contains too many destinations, remove those
// oldest if we did
func (d *destinations) expire() {
	var selected *destination

	if d.length < d.max {
		return
	}

	current := d.pole.Tail

	for removeCount := d.length - d.max; removeCount > 0; removeCount-- {
		if removeCount <= 0 {
			break
		}

		selected = current
		current = selected.prev

		delete(d.destinations, selected.name)
		selected.Delete()

		d.length--
	}
}

// clear removes all destination records
func (d *destinations) clear() {
	var selected *destination

	current := d.pole.Tail

	for {
		if current == nil {
			break
		}

		selected = current
		current = selected.prev

		delete(d.destinations, selected.name)
		selected.Delete()

		d.length--
	}
}

// Get returns transports for specified destination
func (d *destinations) Get(name string) *transports {
	d.destLock.Lock()
	defer d.destLock.Unlock()

	target, found := d.destinations[name]

	if found {
		target.Bump()

		return target.transports
	}

	newDest := &destination{
		name:       name,
		pole:       &d.pole,
		transports: d.build(),
		next:       nil,
		prev:       nil,
	}

	newDest.Attach()

	d.destinations[name] = newDest
	d.length++

	d.expire()

	return newDest.transports
}

// Clear tears down current destinations and clear data in it
func (d *destinations) Clear() {
	d.destLock.Lock()
	defer d.destLock.Unlock()

	d.transports.Kickoff()

	d.clear()
}
