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

package clients

import (
	"sync"

	"github.com/nickrio/coward/roles/common/network/transporter"
)

// Clients is an automatic transporter client selecter
type Clients interface {
	Get(id int) (Client, error)
	Export(callback func(cc Client))
	Kickoff()
	Length() int
}

// clients implements Clients
type clients struct {
	pole    pole
	clients []*client
	length  int
}

// New creates a new Clients
func New(transporters []transporter.Client) Clients {
	var lastClient *client

	tspLength := len(transporters)
	cs := &clients{
		pole: pole{
			Head:     nil,
			Tail:     nil,
			Exported: 0,
			Len:      tspLength,
			Lock:     sync.RWMutex{},
		},
		clients: make([]*client, tspLength),
		length:  tspLength,
	}

	if tspLength > 0 {
		for idx, tsp := range transporters {
			lastClient = newClient(
				idx, lastClient, tsp, &cs.pole)

			cs.clients[idx] = lastClient
		}

		cs.pole.Head = cs.clients[0]
	}

	return cs
}

// iterate iterates client array and hand over array elements
// to a callback function
func (c *clients) iterate(start int, end int, callback func(cc *client) bool) {
	continueIterate := false

	for _, cc := range c.clients[start:end] {
		continueIterate = callback(cc)

		if continueIterate {
			continue
		}

		break
	}
}

// sorted returns a sorted client list. Fast client first
func (c *clients) sorted(callback func(cc *client) bool) {
	current := c.pole.Head
	continueIterate := false

	for {
		if current == nil {
			break
		}

		continueIterate = callback(current)

		if !continueIterate {
			break
		}

		current = current.next
	}
}

// Get returns selected client
func (c *clients) Get(id int) (Client, error) {
	if id >= c.length {
		return nil, ErrClientsNotFound
	}

	return c.clients[id], nil
}

// Export exports clients in Fast To Slow order and ensure
// every transporter client gets chance to be used
func (c *clients) Export(callback func(cc Client)) {
	c.pole.Lock.Lock()
	defer c.pole.Lock.Unlock()

	if c.pole.Len > c.pole.Exported {
		c.iterate(c.pole.Exported, c.pole.Len, func(cc *client) bool {
			callback(cc)

			return true
		})

		c.iterate(0, c.pole.Exported, func(cc *client) bool {
			callback(cc)

			return true
		})

		c.pole.Exported++

		return
	}

	c.sorted(func(cc *client) bool {
		callback(cc)

		return true
	})
}

// Length will return current length of registered transporter clients
func (c *clients) Length() int {
	return c.length
}

// Kickoff disconnect all transporter clients from it's server
func (c *clients) Kickoff() {
	kickWait := sync.WaitGroup{}

	kickWait.Add(c.pole.Len)

	c.iterate(0, c.pole.Len, func(c *client) bool {
		go func() {
			kickWait.Done()

			c.Kickoff()
		}()

		return true
	})

	kickWait.Wait()
}
