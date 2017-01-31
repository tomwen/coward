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
	"github.com/nickrio/coward/roles/common/network/transporter/clients"
)

type dummyTSPHandler struct {
	delay         time.Duration
	delayCallback func(time.Duration)
}

func (d *dummyTSPHandler) Handle() error {
	d.delayCallback(d.delay)

	return nil
}

func (d *dummyTSPHandler) Close() error {
	return nil
}
func (d *dummyTSPHandler) Unleash() error {
	return nil
}
func (d *dummyTSPHandler) Error(
	e error) (wantToRetry bool, wantResetTransportConn bool, err error) {
	return false, false, nil
}

type dummyTSPClient struct {
	name    string
	id      int
	used    bool
	delay   float64
	waiting uint64
	weight  float64
	next    *dummyTSPClient
	prev    *dummyTSPClient
}

func (d *dummyTSPClient) ID() int {
	return d.id
}

func (d *dummyTSPClient) Used() bool {
	return d.used
}

func (d *dummyTSPClient) Next() (clients.Client, error) {
	if d.next == nil {
		return nil, clients.ErrClientsNotFound
	}

	return d.next, nil
}

func (d *dummyTSPClient) Previous() (clients.Client, error) {
	if d.prev == nil {
		return nil, clients.ErrClientsNotFound
	}

	return d.prev, nil
}

func (d *dummyTSPClient) Delay() float64 {
	return d.delay
}

func (d *dummyTSPClient) Weight() float64 {
	return d.weight
}

func (d *dummyTSPClient) Request(
	builder transporter.HandlerBuilder,
	option transporter.RequestOption,
) (bool, error) {
	d.used = true

	builder(transporter.HandlerConfig{}).Handle()

	option.Delay("", d.delay, d.waiting)

	return false, nil
}

func testTransportBuildTSPClients(num int) []clients.Client {
	var lastClient *dummyTSPClient

	results := []clients.Client{}
	names := []string{"ClientA", "ClientB", "ClientC", "ClientD"}

	for i := 0; i < num; i++ {
		newClient := &dummyTSPClient{
			name:   names[i],
			id:     i,
			used:   false,
			delay:  0.0,
			weight: 0,
			next:   nil,
			prev:   lastClient,
		}

		results = append(results, newClient)

		if lastClient != nil {
			lastClient.next = newClient
		}

		lastClient = newClient
	}

	return results
}

func TestTransportsRegister(t *testing.T) {
	csNames := []string{}

	cs := testTransportBuildTSPClients(3)

	tsps := &transports{
		pole: transportPole{
			Head: nil,
			Tail: nil,
			Lock: sync.RWMutex{},
		},
		transports: make([]*transport, 3),
		sorted:     make([]*transport, 3),
		sortLock:   sync.RWMutex{},
	}

	for idx, cli := range cs {
		tsps.Register(idx, cli)
	}

	tsps.iterate(func(tsp *transport) bool {
		csNames = append(csNames, tsp.Client.(*dummyTSPClient).name)

		return true
	})

	if strings.Join(csNames, " ") != "ClientA ClientB ClientC" {
		t.Errorf("Failed to expecting register order %s, got %s",
			"ClientA ClientB ClientC", strings.Join(csNames, " "))

		return
	}
}

func TestTransportsReheadWithWeight(t *testing.T) {
	csNames := []string{}
	cs := testTransportBuildTSPClients(3)
	tsps := &transports{
		pole: transportPole{
			Head: nil,
			Tail: nil,
			Lock: sync.RWMutex{},
		},
		transports: make([]*transport, 3),
		sorted:     make([]*transport, 3),
		sortLock:   sync.RWMutex{},
	}

	// In sorted order
	// A
	cs[0].(*dummyTSPClient).used = false
	cs[0].(*dummyTSPClient).delay = 1.0
	cs[0].(*dummyTSPClient).weight = 70

	// B
	cs[1].(*dummyTSPClient).used = false
	cs[1].(*dummyTSPClient).delay = 5.0
	cs[1].(*dummyTSPClient).weight = 80

	// C
	cs[2].(*dummyTSPClient).used = false
	cs[2].(*dummyTSPClient).delay = 10.1
	cs[2].(*dummyTSPClient).weight = 100

	tsps.Register(0, cs[2])
	tsps.Register(1, cs[1])
	tsps.Register(2, cs[0])

	tsps.transports[0].Delay.Add(cs[0].(*dummyTSPClient).delay)
	tsps.transports[0].Weight = cs[0].(*dummyTSPClient).weight +
		(cs[0].(*dummyTSPClient).weight * cs[0].(*dummyTSPClient).delay)

	tsps.transports[1].Delay.Add(cs[1].(*dummyTSPClient).delay)
	tsps.transports[1].Weight = cs[1].(*dummyTSPClient).weight +
		(cs[1].(*dummyTSPClient).weight * cs[1].(*dummyTSPClient).delay)

	tsps.transports[2].Delay.Add(cs[2].(*dummyTSPClient).delay)
	tsps.transports[2].Weight = cs[2].(*dummyTSPClient).weight +
		(cs[2].(*dummyTSPClient).weight * cs[2].(*dummyTSPClient).delay)

	tsps.iterate(func(tsp *transport) bool {
		csNames = append(csNames, tsp.Client.(*dummyTSPClient).name)

		return true
	})

	if strings.Join(csNames, " ") != "ClientC ClientB ClientA" {
		t.Errorf("Failed to initialize test order, expecting %s, got %s",
			"ClientC ClientB ClientA", csNames)

		return
	}

	// Don't sort as no client is been used yet
	csNames = []string{}

	tsps.rehead()

	tsps.iterate(func(tsp *transport) bool {
		csNames = append(csNames, tsp.Client.(*dummyTSPClient).name)

		return true
	})

	if strings.Join(csNames, " ") != "ClientC ClientB ClientA" {
		t.Errorf("Failed to initialize test order, expecting %s, got %s",
			"ClientC ClientB ClientA", csNames)

		return
	}

	// "Use" all clients
	// A
	cs[0].(*dummyTSPClient).used = true

	// B
	cs[1].(*dummyTSPClient).used = true

	// C
	cs[2].(*dummyTSPClient).used = true

	// Resort the result:
	// rehead method will trying to resort the chain according to
	// clients order (Which should be a chain contains all clients
	// in fast to slow order)
	csNames = []string{}

	tsps.rehead()

	tsps.iterate(func(tsp *transport) bool {
		csNames = append(csNames, tsp.Client.(*dummyTSPClient).name)

		return true
	})

	if strings.Join(csNames, " ") != "ClientB ClientC ClientA" {
		t.Errorf("Failed to initialize test order, expecting %s, got %s",
			"ClientB ClientC ClientA", csNames)

		return
	}

	// Resort again, ClientA will be moved forward since current head
	// is ClientB and ClientA is faster than ClientB
	csNames = []string{}

	tsps.rehead()

	tsps.iterate(func(tsp *transport) bool {
		csNames = append(csNames, tsp.Client.(*dummyTSPClient).name)

		return true
	})

	if strings.Join(csNames, " ") != "ClientA ClientB ClientC" {
		t.Errorf("Failed to initialize test order, expecting %s, got %s",
			"ClientA ClientB ClientC", csNames)

		return
	}

	// Resort again, nothing should be changed this time
	csNames = []string{}

	tsps.rehead()

	tsps.iterate(func(tsp *transport) bool {
		csNames = append(csNames, tsp.Client.(*dummyTSPClient).name)

		return true
	})

	if strings.Join(csNames, " ") != "ClientA ClientB ClientC" {
		t.Errorf("Failed to initialize test order, expecting %s, got %s",
			"ClientA ClientB ClientC", csNames)

		return
	}
}

func TestTransportsReheadWithPrediction(t *testing.T) {
	csNames := []string{}
	cs := testTransportBuildTSPClients(3)
	tsps := &transports{
		pole: transportPole{
			Head: nil,
			Tail: nil,
			Lock: sync.RWMutex{},
		},
		transports: make([]*transport, 3),
		sorted:     make([]*transport, 3),
		sortLock:   sync.RWMutex{},
	}

	// In sorted order
	// A
	cs[0].(*dummyTSPClient).used = false
	cs[0].(*dummyTSPClient).delay = 1.0
	cs[0].(*dummyTSPClient).weight = 70

	// B
	cs[1].(*dummyTSPClient).used = false
	cs[1].(*dummyTSPClient).delay = 5.0
	cs[1].(*dummyTSPClient).weight = 80

	// C
	cs[2].(*dummyTSPClient).used = false
	cs[2].(*dummyTSPClient).delay = 10.1
	cs[2].(*dummyTSPClient).weight = 100

	tsps.Register(0, cs[2])
	tsps.Register(1, cs[1])
	tsps.Register(2, cs[0])

	tsps.transports[0].Delay.Add(cs[0].(*dummyTSPClient).delay)
	tsps.transports[0].Weight = 0

	tsps.transports[1].Delay.Add(cs[1].(*dummyTSPClient).delay)
	tsps.transports[1].Weight = 0

	tsps.transports[2].Delay.Add(cs[2].(*dummyTSPClient).delay)
	tsps.transports[2].Weight = 0

	tsps.iterate(func(tsp *transport) bool {
		csNames = append(csNames, tsp.Client.(*dummyTSPClient).name)

		return true
	})

	if strings.Join(csNames, " ") != "ClientC ClientB ClientA" {
		t.Errorf("Failed to initialize test order, expecting %s, got %s",
			"ClientC ClientB ClientA", csNames)

		return
	}

	// Don't sort as no client is been used yet
	csNames = []string{}

	tsps.rehead()

	tsps.iterate(func(tsp *transport) bool {
		csNames = append(csNames, tsp.Client.(*dummyTSPClient).name)

		return true
	})

	if strings.Join(csNames, " ") != "ClientC ClientB ClientA" {
		t.Errorf("Failed to initialize test order, expecting %s, got %s",
			"ClientC ClientB ClientA", csNames)

		return
	}

	// "Use" all clients
	// A
	cs[0].(*dummyTSPClient).used = true

	// B
	cs[1].(*dummyTSPClient).used = true

	// C
	cs[2].(*dummyTSPClient).used = true

	// Test 1: ClientC is slower than ClientB, so ClientB
	// should be moved to the front of ClientC
	csNames = []string{}

	tsps.rehead()

	tsps.iterate(func(tsp *transport) bool {
		csNames = append(csNames, tsp.Client.(*dummyTSPClient).name)

		return true
	})

	if strings.Join(csNames, " ") != "ClientB ClientC ClientA" {
		t.Errorf("Failed to initialize test order, expecting %s, got %s",
			"ClientB ClientC ClientA", csNames)

		return
	}

	// Test 2: ClientB is slower than ClientA, so ClientA
	// should be moved to the front of ClientB
	csNames = []string{}

	tsps.rehead()

	tsps.iterate(func(tsp *transport) bool {
		csNames = append(csNames, tsp.Client.(*dummyTSPClient).name)

		return true
	})

	if strings.Join(csNames, " ") != "ClientA ClientB ClientC" {
		t.Errorf("Failed to initialize test order, expecting %s, got %s",
			"ClientA ClientB ClientC", csNames)

		return
	}

	// Test 3: Resort again. ClientA is the fastest client so
	// nothing should be changed
	csNames = []string{}

	tsps.rehead()

	tsps.iterate(func(tsp *transport) bool {
		csNames = append(csNames, tsp.Client.(*dummyTSPClient).name)

		return true
	})

	if strings.Join(csNames, " ") != "ClientA ClientB ClientC" {
		t.Errorf("Failed to initialize test order, expecting %s, got %s",
			"ClientA ClientB ClientC", csNames)

		return
	}
}

type dummyPauseableTSPClient struct {
	name      string
	id        int
	used      bool
	delay     float64
	waiting   uint64
	weight    float64
	pauseTime time.Duration
	next      *dummyPauseableTSPClient
	prev      *dummyPauseableTSPClient
}

func (d *dummyPauseableTSPClient) ID() int {
	time.Sleep(d.pauseTime)

	return d.id
}

func (d *dummyPauseableTSPClient) Used() bool {
	time.Sleep(d.pauseTime)

	return d.used
}

func (d *dummyPauseableTSPClient) Next() (clients.Client, error) {
	time.Sleep(d.pauseTime)

	if d.next == nil {
		return nil, clients.ErrClientsNotFound
	}

	return d.next, nil
}

func (d *dummyPauseableTSPClient) Previous() (clients.Client, error) {
	time.Sleep(d.pauseTime)

	if d.prev == nil {
		return nil, clients.ErrClientsNotFound
	}

	return d.prev, nil
}

func (d *dummyPauseableTSPClient) Delay() float64 {
	time.Sleep(d.pauseTime)

	return d.delay
}

func (d *dummyPauseableTSPClient) Weight() float64 {
	time.Sleep(d.pauseTime)

	return d.weight
}

func (d *dummyPauseableTSPClient) Request(
	builder transporter.HandlerBuilder,
	option transporter.RequestOption,
) (bool, error) {
	time.Sleep(d.pauseTime)

	d.used = true

	builder(transporter.HandlerConfig{}).Handle()

	option.Delay("", d.delay, d.waiting)

	return false, nil
}

func TestTransportsRequestFirstRequestMustBlockRest(t *testing.T) {
	failed := false
	wait := sync.WaitGroup{}
	newClient := &dummyPauseableTSPClient{
		name:      "dummyPauseableTSPClient",
		id:        0,
		used:      false,
		delay:     0.0,
		weight:    0,
		pauseTime: time.Duration(0),
		next:      nil,
		prev:      nil,
	}
	tsps := &transports{
		pole: transportPole{
			Head: nil,
			Tail: nil,
			Lock: sync.RWMutex{},
		},
		transports: make([]*transport, 1),
		sorted:     make([]*transport, 1),
		sortLock:   sync.RWMutex{},
		requesting: common.NewCounter(0),
	}

	tsps.Register(0, newClient)

	for i := 0; i < 100; i++ {
		wait.Add(101)

		// Let the first client block the request rehead
		newClient.pauseTime = 60 * time.Millisecond

		go func() {
			defer wait.Done()

			tsps.Request(func(
				cfg transporter.HandlerConfig,
				callback func(time.Duration),
			) transporter.Handler {
				return &dummyTSPHandler{
					delay:         0,
					delayCallback: callback,
				}
			}, transporter.RequestOption{
				Delay: func(addr string, connectDelay float64, waiting uint64) {

				},
			})
		}()

		time.Sleep(10 * time.Millisecond)

		// Not block
		newClient.pauseTime = 0

		for i := 0; i < 100; i++ {
			go func() {
				startTime := time.Now()

				defer func() {
					finished := time.Now().Sub(startTime)

					if finished < 40*time.Millisecond {
						failed = true
					}

					wait.Done()
				}()

				tsps.Request(func(
					cfg transporter.HandlerConfig,
					callback func(time.Duration),
				) transporter.Handler {
					return &dummyTSPHandler{
						delay:         0,
						delayCallback: callback,
					}
				}, transporter.RequestOption{
					Delay: func(
						addr string, connectDelay float64, waiting uint64) {
					},
				})
			}()
		}

		wait.Wait()
	}

	if failed {
		t.Errorf("The first request seems failed to block rest ones")

		return
	}
}

func BenchmarkTransportsRequest100Requests(b *testing.B) {
	wait := sync.WaitGroup{}
	newClient := &dummyPauseableTSPClient{
		name:      "dummyPauseableTSPClient",
		id:        0,
		used:      false,
		delay:     0.0,
		weight:    0,
		pauseTime: time.Duration(0),
		next:      nil,
		prev:      nil,
	}
	tsps := &transports{
		pole: transportPole{
			Head: nil,
			Tail: nil,
			Lock: sync.RWMutex{},
		},
		transports: make([]*transport, 1),
		sorted:     make([]*transport, 1),
		sortLock:   sync.RWMutex{},
		requesting: common.NewCounter(0),
	}

	tsps.Register(0, newClient)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		wait.Add(100)

		for j := 0; j < 100; j++ {
			go func() {
				defer wait.Done()

				tsps.Request(func(
					cfg transporter.HandlerConfig,
					callback func(time.Duration),
				) transporter.Handler {
					return &dummyTSPHandler{
						delay:         0,
						delayCallback: callback,
					}
				}, transporter.RequestOption{
					Delay: func(
						addr string, connectDelay float64, waiting uint64) {
					},
				})
			}()
		}

		wait.Wait()
	}
}
