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

import "github.com/nickrio/coward/roles/common/network/transporter"

// Client is a transporter client re-wrapper
type Client interface {
	ID() int
	Used() bool
	Next() (Client, error)
	Previous() (Client, error)
	Delay() float64
	Weight() float64
	Request(
		builder transporter.HandlerBuilder,
		option transporter.RequestOption,
	) (bool, error)
}

// client implements Client
type client struct {
	id     int
	client transporter.Client
	pole   *pole
	used   bool
	delay  float64
	weight float64
	prev   *client
	next   *client
}

// newClient wraps a transporter client to a chain
func newClient(
	id int,
	appendTo *client,
	t transporter.Client,
	pole *pole,
) *client {
	c := &client{
		id:     id,
		client: t,
		pole:   pole,
		used:   false,
		delay:  0.0,
		weight: 0.0,
		prev:   nil,
		next:   nil,
	}

	if appendTo != nil {
		c.after(appendTo)
	}

	return c
}

// unchain removes the client from chain
func (c *client) unchain() {
	if c == c.pole.Head {
		c.pole.Head = c.next
	}

	if c == c.pole.Tail {
		c.pole.Tail = c.prev
	}

	if c.prev != nil {
		c.prev.next = c.next
	}

	if c.next != nil {
		c.next.prev = c.prev
	}
}

// after puts current client after the target client
func (c *client) after(target *client) {
	c.unchain()

	oldNext := target.next

	target.next = c

	c.next = oldNext
	c.prev = target

	if c.next != nil {
		c.next.prev = c
	} else {
		c.pole.Tail = c
	}
}

// before puts current client before target client
func (c *client) before(target *client) {
	c.unchain()

	oldPrevious := target.prev

	target.prev = c

	c.next = target
	c.prev = oldPrevious

	if c.prev != nil {
		c.prev.next = c
	} else {
		c.pole.Head = c
	}
}

// update updates information about current client and move client's
// position on chain according to it's weight
func (c *client) update(
	connectDelay float64, waiting uint64) {
	c.pole.Lock.Lock()
	defer c.pole.Lock.Unlock()

	c.used = true
	c.weight = connectDelay + (connectDelay * float64(waiting))
	c.delay = connectDelay

	if c.pole.Len > c.pole.Exported {
		return
	}

	if c.next != nil && (c.next.used && c.weight > c.next.weight) {
		selectedClient := c.next
		nextClient := selectedClient

		for {
			if nextClient == nil || nextClient.used {
				break
			}

			if c.weight > nextClient.weight {
				selectedClient = nextClient
				nextClient = selectedClient.next

				continue
			}

			break
		}

		c.after(selectedClient)

		return
	}

	if c.prev != nil && (!c.prev.used || c.weight < c.prev.weight) {
		selectedClient := c.prev
		prevClient := selectedClient

		for {
			if prevClient == nil {
				break
			}

			if !prevClient.used || c.weight < prevClient.weight {
				selectedClient = prevClient
				prevClient = selectedClient.prev

				continue
			}

			break
		}

		c.before(selectedClient)

		return
	}

	return
}

// ID returns the ID of current transporter client
func (c *client) ID() int {
	return c.id
}

// Used return a boolean value to indicate whether or not current
// client has been used before
func (c *client) Used() bool {
	c.pole.Lock.RLock()
	defer c.pole.Lock.RUnlock()

	return c.used
}

// Next returns the next client on chain
func (c *client) Next() (Client, error) {
	c.pole.Lock.RLock()
	defer c.pole.Lock.RUnlock()

	if c.next == nil {
		return nil, ErrClientsNotFound
	}

	return c.next, nil
}

// Next previous returns previous client on chain
func (c *client) Previous() (Client, error) {
	c.pole.Lock.RLock()
	defer c.pole.Lock.RUnlock()

	if c.prev == nil {
		return nil, ErrClientsNotFound
	}

	return c.prev, nil
}

// Request sends a request to the underlaying transporter
func (c *client) Request(
	builder transporter.HandlerBuilder,
	option transporter.RequestOption,
) (bool, error) {
	requested, reqErr := c.client.Request(builder, transporter.RequestOption{
		Buffer:    option.Buffer,
		Canceller: option.Canceller,
		Delay: func(connectDelay float64, waiting uint64) {
			c.update(connectDelay, waiting)

			option.Delay(connectDelay, waiting)
		},
		Error: option.Error,
	})

	return requested, reqErr
}

// Delay returns current connect delay of the transport client
func (c *client) Delay() float64 {
	c.pole.Lock.RLock()
	defer c.pole.Lock.RUnlock()

	return c.delay
}

// Weight returns current weight of the transport client (Light is better)
func (c *client) Weight() float64 {
	c.pole.Lock.RLock()
	defer c.pole.Lock.RUnlock()

	return c.weight
}

// Kickoff drops all transporter connections from it's server
func (c *client) Kickoff() {
	c.client.Kickoff()
}
