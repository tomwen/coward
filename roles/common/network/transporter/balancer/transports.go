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
	"errors"
	"sync"

	"github.com/nickrio/coward/common"
	"github.com/nickrio/coward/roles/common/network/transporter"
	"github.com/nickrio/coward/roles/common/network/transporter/clients"
)

// Transports errors
var (
	ErrTansportsNoTransport = errors.New(
		"No transport to handle your request")
)

// transportPole contains data shared by both transport and transports
// struct
type transportPole struct {
	Head *transport
	Tail *transport
	Lock sync.RWMutex
}

// transports is a chain of transport
type transports struct {
	pole       transportPole
	transports []*transport
	sorted     []*transport
	sortLock   sync.RWMutex
	requesting common.Counter
}

// iterate iterates through all transports
func (t *transports) iterate(callback func(tsp *transport) bool) {
	continueLoop := false
	current := t.pole.Head

	for {
		if current == nil {
			break
		}

		continueLoop = callback(current)

		if !continueLoop {
			break
		}

		current = current.next
	}
}

// rehead select a new head if there is faster client
func (t *transports) rehead() {
	newHead, needsRehead := func() (*transport, bool) {
		if !t.pole.Head.Client.Used() {
			return nil, false
		}

		t.pole.Lock.RLock()
		defer t.pole.Lock.RUnlock()

		betterClient, clientErr := t.pole.Head.Client.Previous()

		if clientErr != nil {
			return nil, false
		}

		if !betterClient.Used() {
			return nil, false
		}

		betterTransport := t.transports[betterClient.ID()]

		// If the betterTransport is known better than current head
		// weight == Transport Connect Delay + Target Connect Delay
		// Smaller weight is better
		if betterTransport.Weight > 0 {
			if betterTransport.Weight > t.pole.Head.Weight {
				return nil, false
			}

			return betterTransport, true
		}

		// If the betterTransport is not been used before, make a
		// prediction about it's speed:
		totalHeadDelay :=
			t.pole.Head.Client.Delay() + t.pole.Head.Delay.Get()

		if betterClient.Delay() > totalHeadDelay {
			return nil, false
		}

		return betterTransport, true
	}()

	if !needsRehead {
		return
	}

	t.pole.Lock.Lock()
	defer t.pole.Lock.Unlock()

	newHead.Before(t.pole.Head)

	t.sortLock.Lock()
	defer t.sortLock.Unlock()

	currentIndex := 0

	t.iterate(func(tsp *transport) bool {
		t.sorted[currentIndex] = tsp

		currentIndex++

		return true
	})
}

// Register adds new client to current transport list
func (t *transports) Register(idx int, c clients.Client) {
	tsp := &transport{
		Client: c,
		Weight: 0,
		Delay:  common.NewLockedAverager(3),
		pole:   &t.pole,
		next:   nil,
		prev:   nil,
	}

	t.transports[c.ID()] = tsp
	t.sorted[idx] = tsp

	if t.pole.Head == nil {
		t.pole.Head = tsp
	}

	if t.pole.Tail == nil {
		t.pole.Tail = tsp

		return
	}

	tsp.After(t.pole.Tail)
}

// Request sends request to server
func (t *transports) Request(
	builder DelayFeedingbackRequestBuilder,
	option transporter.RequestOption,
) error {
	var requestCount uint
	var requested bool
	var reqErr error

	// Only do rehead when there is no request is currently going on
	// to the selected target
	t.requesting.LoadThenAdd(func(count uint64) {
		if count > 0 {
			return
		}

		t.rehead()
	}, 1)
	defer t.requesting.Remove(1)

	t.sortLock.RLock()
	defer t.sortLock.RUnlock()

	for tIdx := range t.sorted {
		requestCount++

		requested, reqErr = t.sorted[tIdx].Request(builder, option)

		if !requested {
			continue
		}

		break
	}

	if requestCount <= 0 {
		return ErrTansportsNoTransport
	}

	return reqErr
}
