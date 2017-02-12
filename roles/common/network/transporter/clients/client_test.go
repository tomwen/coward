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
	"math/rand"
	"strings"
	"sync"
	"testing"

	"github.com/nickrio/coward/roles/common/network/transporter"
)

type dummyClient struct {
	id string
}

func (d *dummyClient) Request(
	builder transporter.HandlerBuilder,
	option transporter.RequestOption) (bool, error) {
	return true, nil
}

func (d *dummyClient) Kickoff() {}

type dummyDelayedClient struct {
	id      string
	waiting uint64
	delay   float64
}

func (d *dummyDelayedClient) Request(
	builder transporter.HandlerBuilder,
	option transporter.RequestOption) (bool, error) {
	option.Delay(d.delay, d.waiting)

	return true, nil
}

func (d *dummyDelayedClient) Kickoff() {}

type dummyRandomDelayedClient struct {
	id      string
	waiting uint64
	delay   float64
}

func (d *dummyRandomDelayedClient) Request(
	builder transporter.HandlerBuilder,
	option transporter.RequestOption) (bool, error) {
	d.waiting = uint64(rand.Int63())
	d.delay = rand.Float64()

	option.Delay(d.delay, d.waiting)

	return true, nil
}

func (d *dummyRandomDelayedClient) Kickoff() {}

func testPrioritiesBuildClientLists(num int) []transporter.Client {
	names := []string{
		"ClientA", "ClientB", "ClientC", "ClientD", "ClientE", "ClientF",
		"ClientG", "ClientH", "ClientI", "ClientJ", "ClientK", "ClientL",
		"ClientM", "ClientN", "ClientO", "ClientP", "ClientQ", "ClientI",
		"ClientS", "ClientT", "ClientU", "ClientV", "ClientW", "ClientX",
		"ClientY", "ClientZ",
	}
	clist := []transporter.Client{}

	for i := 0; i < num; i++ {
		clist = append(clist, &dummyClient{
			id: names[i%len(names)],
		})
	}

	return clist
}

func testClientBuildClientChan(num int) *client {
	k := &pole{
		Head: nil,
		Tail: nil,
		Lock: sync.RWMutex{},
	}

	transporters := testPrioritiesBuildClientLists(num)
	c := newClient(0, nil, transporters[0], k)
	sub := c

	k.Head = c

	for idx, transporter := range transporters[1:] {
		sub = newClient(idx+1, sub, transporter, k)
	}

	return c
}

func testClientChainListStringFromHead(head *client) []string {
	result := []string{}

	current := head

	for {
		if current == nil {
			break
		}

		result = append(result, current.client.(*dummyClient).id)

		current = current.next
	}

	return result
}

func testClientChainListStringFromTail(prev *client) []string {
	result := []string{}

	current := prev

	for {
		if current == nil {
			break
		}

		result = append(result, current.client.(*dummyClient).id)

		current = current.prev
	}

	return result
}

func TestClientAfter(t *testing.T) {
	head := testClientBuildClientChan(3)

	clientA := head
	clientB := head.next
	clientC := head.next.next

	// Test 1: Move the middle element to the tail
	clientB.after(clientC)

	if "ClientA ClientC ClientB" != strings.Join(
		testClientChainListStringFromHead(head.pole.Head), " ") {
		t.Errorf("Failed to expecting client order to be %s, got %s",
			"ClientA ClientC ClientB", strings.Join(
				testClientChainListStringFromHead(head.pole.Head), " "))

		return
	}

	if "ClientB ClientC ClientA" != strings.Join(
		testClientChainListStringFromTail(head.pole.Tail), " ") {
		t.Errorf("Failed to expecting client order to be %s, got %s",
			"ClientB ClientC ClientA", strings.Join(
				testClientChainListStringFromTail(head.pole.Tail), " "))

		return
	}

	// Test 2: Move the head element to the tail
	clientA.after(clientB)

	if "ClientC ClientB ClientA" != strings.Join(
		testClientChainListStringFromHead(head.pole.Head), " ") {
		t.Errorf("Failed to expecting client order to be %s, got %s",
			"ClientC ClientB ClientA", strings.Join(
				testClientChainListStringFromHead(head.pole.Head), " "))

		return
	}

	if "ClientA ClientB ClientC" != strings.Join(
		testClientChainListStringFromTail(head.pole.Tail), " ") {
		t.Errorf("Failed to expecting client order to be %s, got %s",
			"ClientA ClientB ClientC", strings.Join(
				testClientChainListStringFromTail(head.pole.Tail), " "))

		return
	}

	// Test 3: Move the middle element to the tail again
	clientB.after(clientA)

	if "ClientC ClientA ClientB" != strings.Join(
		testClientChainListStringFromHead(head.pole.Head), " ") {
		t.Errorf("Failed to expecting client order to be %s, got %s",
			"ClientC ClientA ClientB", strings.Join(
				testClientChainListStringFromHead(head.pole.Head), " "))

		return
	}

	if "ClientB ClientA ClientC" != strings.Join(
		testClientChainListStringFromTail(head.pole.Tail), " ") {
		t.Errorf("Failed to expecting client order to be %s, got %s",
			"ClientB ClientA ClientC", strings.Join(
				testClientChainListStringFromTail(head.pole.Tail), " "))

		return
	}
}

func TestClientBefore(t *testing.T) {
	head := testClientBuildClientChan(3)

	clientA := head
	clientB := head.next
	clientC := head.next.next

	// Test 1: Move the middle element to the head
	clientB.before(clientA)

	if "ClientB ClientA ClientC" != strings.Join(
		testClientChainListStringFromHead(head.pole.Head), " ") {
		t.Errorf("Failed to expecting client order to be %s, got %s",
			"ClientB ClientA ClientC", strings.Join(
				testClientChainListStringFromHead(head.pole.Head), " "))

		return
	}

	if "ClientC ClientA ClientB" != strings.Join(
		testClientChainListStringFromTail(head.pole.Tail), " ") {
		t.Errorf("Failed to expecting client order to be %s, got %s",
			"ClientC ClientA ClientB", strings.Join(
				testClientChainListStringFromTail(head.pole.Tail), " "))

		return
	}

	// Test 2: Move the tail element to the middle
	clientC.before(clientA)

	if "ClientB ClientC ClientA" != strings.Join(
		testClientChainListStringFromHead(head.pole.Head), " ") {
		t.Errorf("Failed to expecting client order to be %s, got %s",
			"ClientB ClientC ClientA", strings.Join(
				testClientChainListStringFromHead(head.pole.Head), " "))

		return
	}

	if "ClientA ClientC ClientB" != strings.Join(
		testClientChainListStringFromTail(head.pole.Tail), " ") {
		t.Errorf("Failed to expecting client order to be %s, got %s",
			"ClientA ClientC ClientB", strings.Join(
				testClientChainListStringFromTail(head.pole.Tail), " "))

		return
	}

	// Test 3: Move the tail element to the head
	clientA.before(clientB)

	if "ClientA ClientB ClientC" != strings.Join(
		testClientChainListStringFromHead(head.pole.Head), " ") {
		t.Errorf("Failed to expecting client order to be %s, got %s",
			"ClientA ClientB ClientC", strings.Join(
				testClientChainListStringFromHead(head.pole.Head), " "))

		return
	}

	if "ClientC ClientB ClientA" != strings.Join(
		testClientChainListStringFromTail(head.pole.Tail), " ") {
		t.Errorf("Failed to expecting client order to be %s, got %s",
			"ClientC ClientB ClientA", strings.Join(
				testClientChainListStringFromTail(head.pole.Tail), " "))

		return
	}

	// Test 4: Move the tail element to the head
	clientC.before(clientA)

	if "ClientC ClientA ClientB" != strings.Join(
		testClientChainListStringFromHead(head.pole.Head), " ") {
		t.Errorf("Failed to expecting client order to be %s, got %s",
			"ClientC ClientA ClientB", strings.Join(
				testClientChainListStringFromHead(head.pole.Head), " "))

		return
	}

	if "ClientB ClientA ClientC" != strings.Join(
		testClientChainListStringFromTail(head.pole.Tail), " ") {
		t.Errorf("Failed to expecting client order to be %s, got %s",
			"ClientB ClientA ClientC", strings.Join(
				testClientChainListStringFromTail(head.pole.Tail), " "))

		return
	}
}

func TestClientRequestOrderFullSort(t *testing.T) {
	sortedOrder := []string{}
	clientA := &dummyDelayedClient{
		id:      "ClientA",
		waiting: 1,
		delay:   101.0,
	}
	clientB := &dummyDelayedClient{
		id:      "ClientB",
		waiting: 0,
		delay:   101.0,
	}
	clientC := &dummyDelayedClient{
		id:      "ClientC",
		waiting: 0,
		delay:   1.0,
	}

	cs := New([]transporter.Client{
		clientA, clientB, clientC,
	})

	cList := []Client{}

	cs.(*clients).iterate(0, cs.(*clients).pole.Len, func(c *client) bool {
		cList = append(cList, c)

		return true
	})

	// After all transporter has been exported
	for _ = range cList {
		cs.Export(func(cc Client) {})
	}

	for _, cl := range cList {
		cl.Request(
			func(transporter.HandlerConfig) transporter.Handler {
				return nil
			},
			transporter.RequestOption{
				Delay: func(delay float64, waiting uint64) {},
			},
		)
	}

	cs.(*clients).sorted(func(c *client) bool {
		sortedOrder = append(
			sortedOrder,
			c.client.(*dummyDelayedClient).id,
		)

		return true
	})

	if strings.Join(sortedOrder, " ") != "ClientC ClientB ClientA" {
		t.Errorf("Expected Client order should be %s, got %s",
			"ClientC ClientB ClientA", strings.Join(sortedOrder, " "))

		return
	}
}

func TestClientRequestOrderSortFirst2(t *testing.T) {
	sortIndex := 0
	sortedOrder := []string{}
	clientA := &dummyDelayedClient{
		id:      "ClientA",
		waiting: 1,
		delay:   101.0,
	}
	clientB := &dummyDelayedClient{
		id:      "ClientB",
		waiting: 0,
		delay:   10.0,
	}
	clientC := &dummyDelayedClient{
		id:      "ClientC",
		waiting: 0,
		delay:   1.0,
	}

	cs := New([]transporter.Client{
		clientA, clientB, clientC,
	})

	cList := []Client{}

	cs.(*clients).iterate(0, cs.(*clients).pole.Len, func(c *client) bool {
		cList = append(cList, c)

		if sortIndex >= 1 {
			return false
		}

		sortIndex++

		return true
	})

	cs.Export(func(cc Client) {})
	cs.Export(func(cc Client) {})
	cs.Export(func(cc Client) {})

	for _, cl := range cList {
		cl.Request(
			func(transporter.HandlerConfig) transporter.Handler {
				return nil
			},
			transporter.RequestOption{
				Delay: func(delay float64, waiting uint64) {},
			},
		)
	}

	cs.(*clients).sorted(func(c *client) bool {
		sortedOrder = append(
			sortedOrder,
			c.client.(*dummyDelayedClient).id,
		)

		return true
	})

	if strings.Join(sortedOrder, " ") != "ClientB ClientA ClientC" {
		t.Errorf("Expected Client order should be %s, got %s",
			"ClientB ClientA ClientC", strings.Join(sortedOrder, " "))

		return
	}
}

func TestClientRequestOrderSortFirst1(t *testing.T) {
	sortedOrder := []string{}
	clientA := &dummyDelayedClient{
		id:      "ClientA",
		waiting: 1,
		delay:   101.0,
	}
	clientB := &dummyDelayedClient{
		id:      "ClientB",
		waiting: 0,
		delay:   10.0,
	}
	clientC := &dummyDelayedClient{
		id:      "ClientC",
		waiting: 0,
		delay:   1.0,
	}

	cs := New([]transporter.Client{
		clientA, clientB, clientC,
	})

	cList := []Client{}

	cs.(*clients).iterate(0, cs.(*clients).pole.Len, func(c *client) bool {
		cList = append(cList, c)

		return false
	})

	for _, cl := range cList {
		cl.Request(
			func(transporter.HandlerConfig) transporter.Handler {
				return nil
			},
			transporter.RequestOption{
				Delay: func(delay float64, waiting uint64) {},
			},
		)
	}

	cs.(*clients).sorted(func(c *client) bool {
		sortedOrder = append(
			sortedOrder,
			c.client.(*dummyDelayedClient).id,
		)

		return true
	})

	if strings.Join(sortedOrder, " ") != "ClientA ClientB ClientC" {
		t.Errorf("Expected Client order should be %s, got %s",
			"ClientA ClientB ClientC", strings.Join(sortedOrder, " "))
	}
}

func BenchmarkClientRequest(b *testing.B) {
	var targetClient Client
	var targetIndex int

	clientA := &dummyRandomDelayedClient{
		id:      "ClientA",
		waiting: 0,
		delay:   0,
	}
	clientB := &dummyRandomDelayedClient{
		id:      "ClientB",
		waiting: 0,
		delay:   0,
	}
	clientC := &dummyRandomDelayedClient{
		id:      "ClientC",
		waiting: 0,
		delay:   0,
	}
	clientD := &dummyRandomDelayedClient{
		id:      "ClientD",
		waiting: 0,
		delay:   0,
	}

	cs := New([]transporter.Client{
		clientA, clientB, clientC, clientD,
	})

	cList := []Client{}

	cs.(*clients).iterate(0, cs.(*clients).pole.Len, func(c *client) bool {
		cList = append(cList, c)

		if targetIndex == 1 {
			targetClient = c

			return false
		}

		targetIndex++

		return true
	})

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		targetClient.Request(
			func(transporter.HandlerConfig) transporter.Handler {
				return nil
			},
			transporter.RequestOption{
				Delay: func(delay float64, waiting uint64) {},
			},
		)
	}
}
