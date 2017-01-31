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
	"time"

	"github.com/nickrio/coward/common"
	"github.com/nickrio/coward/roles/common/network/transporter"
	"github.com/nickrio/coward/roles/common/network/transporter/clients"
)

// transport is wrapped client for better and easier sorting
type transport struct {
	Client clients.Client
	Weight float64
	Delay  common.Averager
	pole   *transportPole
	next   *transport
	prev   *transport
}

// unchain takes out a transport node from chain
func (t *transport) unchain() {
	if t == t.pole.Head {
		t.pole.Head = t.next
	}

	if t == t.pole.Tail {
		t.pole.Tail = t.prev
	}

	if t.prev != nil {
		t.prev.next = t.next
	}

	if t.next != nil {
		t.next.prev = t.prev
	}
}

// reweight updates current transport's weight safely
func (t *transport) reweight(newWeight float64) {
	t.pole.Lock.Lock()
	defer t.pole.Lock.Unlock()

	t.Weight = newWeight
}

// After puts current transport after another transport
func (t *transport) After(target *transport) {
	t.unchain()

	oldNext := target.next

	target.next = t

	t.next = oldNext
	t.prev = target

	if t.next != nil {
		t.next.prev = t
	} else {
		t.pole.Tail = t
	}
}

// Before puts current transport before another transport
func (t *transport) Before(target *transport) {
	t.unchain()

	oldPrevious := target.prev

	target.prev = t

	t.next = target
	t.prev = oldPrevious

	if t.prev != nil {
		t.prev.next = t
	} else {
		t.pole.Head = t
	}
}

// Request sends request to the remove server
func (t *transport) Request(
	builder DelayFeedingbackRequestBuilder,
	option transporter.RequestOption,
) (bool, error) {
	var newDelay float64
	var newWeight float64

	return t.Client.Request(
		func(cfg transporter.HandlerConfig) transporter.Handler {
			return builder(cfg, func(delay time.Duration) {
				newDelay = t.Delay.Add(delay.Seconds())

				if newWeight != 0 {
					t.reweight(newWeight + newDelay)
				}
			})
		}, transporter.RequestOption{
			Buffer:    option.Buffer,
			Canceller: option.Canceller,
			Delay: func(addr string, connectDelay float64, waiting uint64) {
				newWeight =
					connectDelay + (connectDelay * float64(waiting))

				if newDelay != 0.0 {
					t.reweight(newWeight + newDelay)
				}

				option.Delay(addr, connectDelay, waiting)
			},
			Error: option.Error,
		})
}
