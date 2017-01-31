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

	"github.com/nickrio/coward/roles/common/network/transporter"
	"github.com/nickrio/coward/roles/common/network/transporter/clients"
)

// Balancer is an automatic transporter client selector
// which trying to always provide fastest transporter for request to use
type Balancer interface {
	Request(
		string,
		DelayFeedingbackRequestBuilder,
		transporter.RequestOption,
	) error
	Kickoff()
}

// balancer implements Balancer
type balancer struct {
	dests destinations
}

// New creates a new Balancer
func New(c clients.Clients, maxDests uint) Balancer {
	return &balancer{
		dests: destinations{
			max:    maxDests,
			length: 0,
			pole: destinationPole{
				Head: nil,
				Tail: nil,
			},
			transports:   c,
			destinations: make(map[string]*destination, maxDests),
			destLock:     sync.Mutex{},
		},
	}
}

// Request picks up remote connections to handle a request
func (b *balancer) Request(
	target string,
	builder DelayFeedingbackRequestBuilder,
	option transporter.RequestOption,
) error {
	return b.dests.Get(target).Request(builder, option)
}

// Kickoff tears current balancer down
func (b *balancer) Kickoff() {
	b.dests.Clear()
}
