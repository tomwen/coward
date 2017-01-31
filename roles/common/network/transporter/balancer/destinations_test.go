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
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/nickrio/coward/roles/common/network/transporter"
	"github.com/nickrio/coward/roles/common/network/transporter/clients"
)

func TestDestinationsGet(t *testing.T) {
	cli := clients.New([]transporter.Client{})

	d := &destinations{
		max:    5,
		length: 0,
		pole: destinationPole{
			Head: nil,
			Tail: nil,
		},
		transports:   cli,
		destinations: make(map[string]*destination, 5),
		destLock:     sync.Mutex{},
	}

	d.Get("D1")
	d.Get("D2")

	if d.length != 2 {
		t.Error("Failed to update length count")

		return
	}

	if len(d.destinations) != 2 {
		t.Error("Failed to save destination record")

		return
	}

	if strings.Join(testDestinationTraverse(&d.pole), " ") !=
		"D2 D1 D1 D2" {
		t.Errorf("Expecting remaining client will be %s, got %s",
			"D2 D1 D1 D2",
			testDestinationTraverse(&d.pole))

		return
	}

	d.Get("D3")
	d.Get("D4")
	d.Get("D5")
	d.Get("D6")
	d.Get("D7")
	d.Get("D8")

	if d.length != 5 {
		t.Error("Failed to expire old destination record")

		return
	}

	if len(d.destinations) != 5 {
		t.Error("Failed to remove old destination items")

		return
	}

	if strings.Join(testDestinationTraverse(&d.pole), " ") !=
		"D8 D7 D6 D5 D4 D4 D5 D6 D7 D8" {
		t.Errorf("Expecting remaining client will be %s, got %s",
			"D8 D7 D6 D5 D4 D4 D5 D6 D7 D8",
			testDestinationTraverse(&d.pole))

		return
	}

	d.Clear()

	if d.length != 0 {
		t.Error("Failed to expire old destination record")

		return
	}

	if len(d.destinations) != 0 {
		t.Error("Failed to remove old destination items")

		return
	}

	if strings.Join(testDestinationTraverse(&d.pole), " ") != "" {
		t.Errorf("Expecting remaining client will be empty, got %s",
			testDestinationTraverse(&d.pole))

		return
	}

	d.Get("D1")
	d.Get("D2")

	if d.length != 2 {
		t.Error("Failed to update length count")

		return
	}

	if len(d.destinations) != 2 {
		t.Error("Failed to save destination record")

		return
	}

	if strings.Join(testDestinationTraverse(&d.pole), " ") !=
		"D2 D1 D1 D2" {
		t.Errorf("Expecting remaining client will be %s, got %s",
			"D2 D1 D1 D2",
			testDestinationTraverse(&d.pole))

		return
	}
}

func TestDestinationsGetExpire(t *testing.T) {
	cli := clients.New([]transporter.Client{})

	d := &destinations{
		max:    5,
		length: 0,
		pole: destinationPole{
			Head: nil,
			Tail: nil,
		},
		transports:   cli,
		destinations: make(map[string]*destination, 5),
		destLock:     sync.Mutex{},
	}

	for i := 0; i < 20; i++ {
		d.Get(fmt.Sprintf("D%d", i))
	}

	if strings.Join(testDestinationTraverse(&d.pole), " ") !=
		"D19 D18 D17 D16 D15 D15 D16 D17 D18 D19" {
		t.Errorf("Expecting remaining client will be %s, got %s",
			"D19 D18 D17 D16 D15 D15 D16 D17 D18 D19",
			testDestinationTraverse(&d.pole))

		return
	}
}
