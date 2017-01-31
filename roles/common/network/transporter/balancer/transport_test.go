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
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/nickrio/coward/common"
	"github.com/nickrio/coward/roles/common/network/transporter"
)

func testTransportGenerateClients(num int) []*transport {
	var lastTSP *transport

	clients := testTransportBuildTSPClients(3)
	pole := &transportPole{
		Head: nil,
		Tail: nil,
		Lock: sync.RWMutex{},
	}
	transports := []*transport{}

	for _, c := range clients {
		newTSP := &transport{
			Client: c,
			Weight: 0,
			Delay:  common.NewLockedAverager(3),
			pole:   pole,
			next:   nil,
			prev:   lastTSP,
		}

		transports = append(transports, newTSP)

		if pole.Head == nil {
			pole.Head = newTSP
		}

		if lastTSP != nil {
			lastTSP.next = newTSP
		}

		lastTSP = newTSP
	}

	pole.Tail = lastTSP

	return transports
}

func testTransportTraverseChainFromHead(p *transportPole) []string {
	clientNames := []string{}
	current := p.Head

	for {
		if current == nil {
			break
		}

		clientNames = append(
			clientNames, current.Client.(*dummyTSPClient).name)

		current = current.next
	}

	return clientNames
}

func testTransportTraverseChainFromTail(p *transportPole) []string {
	clientNames := []string{}
	current := p.Tail

	for {
		if current == nil {
			break
		}

		clientNames = append(
			clientNames, current.Client.(*dummyTSPClient).name)

		current = current.prev
	}

	return clientNames
}

func TestTransportAfter(t *testing.T) {
	tsps := testTransportGenerateClients(3)

	if strings.Join(
		testTransportTraverseChainFromHead(tsps[0].pole),
		" ",
	) != "ClientA ClientB ClientC" {
		t.Errorf("Failed to initialize test order. Expecting %s, got %s",
			"ClientA ClientB ClientC",
			testTransportTraverseChainFromHead(tsps[0].pole))

		return
	}

	if strings.Join(
		testTransportTraverseChainFromTail(tsps[0].pole),
		" ",
	) != "ClientC ClientB ClientA" {
		t.Errorf("Failed to initialize test order. Expecting %s, got %s",
			"ClientC ClientB ClientA",
			testTransportTraverseChainFromTail(tsps[0].pole))

		return
	}

	// Test 1: Move a middle element to the tail
	tsps[1].After(tsps[2])

	if strings.Join(
		testTransportTraverseChainFromHead(tsps[0].pole),
		" ",
	) != "ClientA ClientC ClientB" {
		t.Errorf("Failed to move the element. Expecting %s, got %s",
			"ClientA ClientC ClientB",
			testTransportTraverseChainFromHead(tsps[0].pole))

		return
	}

	if strings.Join(
		testTransportTraverseChainFromTail(tsps[0].pole),
		" ",
	) != "ClientB ClientC ClientA" {
		t.Errorf("Failed to move the element. Expecting %s, got %s",
			"ClientB ClientC ClientA",
			testTransportTraverseChainFromTail(tsps[0].pole))

		return
	}

	// Test 2: Move a head element to the middle
	tsps[0].After(tsps[2])

	if strings.Join(
		testTransportTraverseChainFromHead(tsps[0].pole),
		" ",
	) != "ClientC ClientA ClientB" {
		t.Errorf("Failed to move the element. Expecting %s, got %s",
			"ClientC ClientA ClientB",
			testTransportTraverseChainFromHead(tsps[0].pole))

		return
	}

	if strings.Join(
		testTransportTraverseChainFromTail(tsps[0].pole),
		" ",
	) != "ClientB ClientA ClientC" {
		t.Errorf("Failed to move the element. Expecting %s, got %s",
			"ClientB ClientA ClientC",
			testTransportTraverseChainFromTail(tsps[0].pole))

		return
	}

	// Test 3: Move a head element to the tail
	tsps[2].After(tsps[1])

	if strings.Join(
		testTransportTraverseChainFromHead(tsps[0].pole),
		" ",
	) != "ClientA ClientB ClientC" {
		t.Errorf("Failed to move the element. Expecting %s, got %s",
			"ClientA ClientB ClientC",
			testTransportTraverseChainFromHead(tsps[0].pole))

		return
	}

	if strings.Join(
		testTransportTraverseChainFromTail(tsps[0].pole),
		" ",
	) != "ClientC ClientB ClientA" {
		t.Errorf("Failed to move the element. Expecting %s, got %s",
			"ClientC ClientB ClientA",
			testTransportTraverseChainFromTail(tsps[0].pole))

		return
	}
}

func TestTransportBefore(t *testing.T) {
	tsps := testTransportGenerateClients(3)

	if strings.Join(
		testTransportTraverseChainFromHead(tsps[0].pole),
		" ",
	) != "ClientA ClientB ClientC" {
		t.Errorf("Failed to initialize test order. Expecting %s, got %s",
			"ClientA ClientB ClientC",
			testTransportTraverseChainFromHead(tsps[0].pole))

		return
	}

	if strings.Join(
		testTransportTraverseChainFromTail(tsps[0].pole),
		" ",
	) != "ClientC ClientB ClientA" {
		t.Errorf("Failed to initialize test order. Expecting %s, got %s",
			"ClientC ClientB ClientA",
			testTransportTraverseChainFromTail(tsps[0].pole))

		return
	}

	// Test 1: Move a middle element to the head
	tsps[1].Before(tsps[0])

	if strings.Join(
		testTransportTraverseChainFromHead(tsps[0].pole),
		" ",
	) != "ClientB ClientA ClientC" {
		t.Errorf("Failed to move the element. Expecting %s, got %s",
			"ClientB ClientA ClientC",
			testTransportTraverseChainFromHead(tsps[0].pole))

		return
	}

	if strings.Join(
		testTransportTraverseChainFromTail(tsps[0].pole),
		" ",
	) != "ClientC ClientA ClientB" {
		t.Errorf("Failed to move the element. Expecting %s, got %s",
			"ClientC ClientA ClientB",
			testTransportTraverseChainFromTail(tsps[0].pole))

		return
	}

	// Test 2: Move a tail element to the middle
	tsps[2].Before(tsps[0])

	if strings.Join(
		testTransportTraverseChainFromHead(tsps[0].pole),
		" ",
	) != "ClientB ClientC ClientA" {
		t.Errorf("Failed to move the element. Expecting %s, got %s",
			"ClientB ClientC ClientA",
			testTransportTraverseChainFromHead(tsps[0].pole))

		return
	}

	if strings.Join(
		testTransportTraverseChainFromTail(tsps[0].pole),
		" ",
	) != "ClientA ClientC ClientB" {
		t.Errorf("Failed to move the element. Expecting %s, got %s",
			"ClientA ClientC ClientB",
			testTransportTraverseChainFromTail(tsps[0].pole))

		return
	}

	// Test 3: Move a tail element to the head
	tsps[0].Before(tsps[1])

	if strings.Join(
		testTransportTraverseChainFromHead(tsps[0].pole),
		" ",
	) != "ClientA ClientB ClientC" {
		t.Errorf("Failed to move the element. Expecting %s, got %s",
			"ClientA ClientB ClientC",
			testTransportTraverseChainFromHead(tsps[0].pole))

		return
	}

	if strings.Join(
		testTransportTraverseChainFromTail(tsps[0].pole),
		" ",
	) != "ClientC ClientB ClientA" {
		t.Errorf("Failed to move the element. Expecting %s, got %s",
			"ClientC ClientB ClientA",
			testTransportTraverseChainFromTail(tsps[0].pole))

		return
	}
}

func dummyDelayFeedingBackRequestBuilderBuilder(
	delay time.Duration) DelayFeedingbackRequestBuilder {
	return func(
		cfg transporter.HandlerConfig,
		callback func(time.Duration),
	) transporter.Handler {
		return &dummyTSPHandler{
			delay:         delay,
			delayCallback: callback,
		}
	}
}

func dummyDelayFeedingBackRequestBuilder(
	cfg transporter.HandlerConfig,
	callback func(time.Duration),
) transporter.Handler {
	return &dummyTSPHandler{
		delay:         0,
		delayCallback: callback,
	}
}

func TestTransportRequest(t *testing.T) {
	cli := &dummyTSPClient{
		name:    "Client",
		id:      0,
		used:    false,
		delay:   2,
		waiting: 5,
		weight:  20,
		next:    nil,
		prev:    nil,
	}
	tsp := &transport{
		Client: cli,
		Weight: 0,
		Delay:  common.NewLockedAverager(3),
		pole: &transportPole{
			Lock: sync.RWMutex{},
		},
		next: nil,
		prev: nil,
	}

	if tsp.Weight != 0 {
		t.Error("Blank transport must have 0 weight")

		return
	}

	// Test 1:
	tsp.Request(
		dummyDelayFeedingBackRequestBuilderBuilder(3*time.Second),
		transporter.RequestOption{
			Delay: func(addr string, connectDelay float64, waiting uint64) {},
		},
	)

	// 3.0 / 1 = 3
	if tsp.Delay.Get() != 3.0 {
		t.Errorf("Expecting delay to be %f, got %f",
			3.0, tsp.Delay.Get())

		return
	}

	// 2 + (2 * 5) + 3 = 15
	if tsp.Weight != 15.0 {
		t.Errorf("Failed to sum up weight. Expecting %f, got %f",
			15.0, tsp.Weight)

		return
	}

	// Test 2:
	cli.delay = 5
	cli.waiting = 10

	tsp.Request(
		dummyDelayFeedingBackRequestBuilderBuilder(30*time.Second),
		transporter.RequestOption{
			Delay: func(addr string, connectDelay float64, waiting uint64) {},
		},
	)

	// 30 + 3 / 2
	if tsp.Delay.Get() != 16.5 {
		t.Errorf("Expecting delay to be %f, got %f",
			16.5, tsp.Delay.Get())

		return
	}

	// 5 + (5 * 10) + 16.5 = 71
	if tsp.Weight != 71.5 {
		t.Errorf("Failed to sum up weight. Expecting %f, got %f",
			71.5, tsp.Weight)

		return
	}
}

func BenchmarkTransportRequest(b *testing.B) {
	cli := &dummyTSPClient{
		name:    "Client",
		id:      0,
		used:    false,
		delay:   2,
		waiting: 5,
		weight:  20,
		next:    nil,
		prev:    nil,
	}
	tsp := &transport{
		Client: cli,
		Weight: 0,
		Delay:  common.NewLockedAverager(3),
		pole: &transportPole{
			Lock: sync.RWMutex{},
		},
		next: nil,
		prev: nil,
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tsp.Request(
			dummyDelayFeedingBackRequestBuilder,
			transporter.RequestOption{
				Delay: func(addr string, connectDelay float64, waiting uint64) {

				},
			},
		)
	}
}
