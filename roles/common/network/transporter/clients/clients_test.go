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
	"strings"
	"testing"

	"github.com/nickrio/coward/roles/common/network/transporter"
)

func TestClientsExport(t *testing.T) {
	clientOrder := []string{}
	clientA := &dummyDelayedClient{
		id:      "ClientA",
		waiting: 0,
		delay:   101.0,
	}
	clientB := &dummyDelayedClient{
		id:      "ClientB",
		waiting: 1,
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

	// Export for the first time, order should be
	// ClientA ClientB ClientC
	cs.Export(func(c Client) {
		clientOrder = append(
			clientOrder, c.(*client).client.(*dummyDelayedClient).id)
	})

	if strings.Join(clientOrder, " ") != "ClientA ClientB ClientC" {
		t.Errorf("Failed to export client in expected order %s, got %s",
			"ClientA ClientB ClientC", clientOrder)

		return
	}

	// Test 1: Use the client 1, see if the exported client changes
	// it's order.
	// It shouldn't, so the result should be:
	// ClientB ClientC ClientA (rolled)
	cs.(*clients).clients[0].Request(
		func(transporter.HandlerConfig) transporter.Handler {
			return nil
		},
		transporter.RequestOption{
			Delay: func(addr string, delay float64, waiting uint64) {},
		},
	)

	clientOrder = []string{}

	cs.Export(func(c Client) {
		clientOrder = append(
			clientOrder, c.(*client).client.(*dummyDelayedClient).id)
	})

	if strings.Join(clientOrder, " ") != "ClientB ClientC ClientA" {
		t.Errorf("Failed to export client in expected order %s, got %s",
			"ClientB ClientC ClientA", clientOrder)

		return
	}

	// Test 2: Use the client 2, see if the exported client changes
	// it's order.
	// It shouldn't, so the result should be:
	// ClientC ClientA ClientB (rolled)
	cs.(*clients).clients[1].Request(
		func(transporter.HandlerConfig) transporter.Handler {
			return nil
		},
		transporter.RequestOption{
			Delay: func(addr string, delay float64, waiting uint64) {},
		},
	)

	clientOrder = []string{}

	cs.Export(func(c Client) {
		clientOrder = append(
			clientOrder, c.(*client).client.(*dummyDelayedClient).id)
	})

	if strings.Join(clientOrder, " ") != "ClientC ClientA ClientB" {
		t.Errorf("Failed to export client in expected order %s, got %s",
			"ClientC ClientA ClientB", clientOrder)

		return
	}

	// Test 3: Use the client 3, see if the exported client changes
	// it's order.
	// It should, so the result should be as sorted:
	// ClientC ClientA ClientB
	cs.(*clients).clients[2].Request(
		func(transporter.HandlerConfig) transporter.Handler {
			return nil
		},
		transporter.RequestOption{
			Delay: func(addr string, delay float64, waiting uint64) {},
		},
	)

	clientOrder = []string{}

	cs.Export(func(c Client) {
		clientOrder = append(
			clientOrder, c.(*client).client.(*dummyDelayedClient).id)
	})

	if strings.Join(clientOrder, " ") != "ClientC ClientA ClientB" {
		t.Errorf("Failed to export client in expected order %s, got %s",
			"ClientC ClientA ClientB", clientOrder)

		return
	}
}

func BenchmarkClientsExport(b *testing.B) {
	clientsList := make([]Client, 4)
	clientA := &dummyDelayedClient{
		id:      "ClientA",
		waiting: 0,
		delay:   101.0,
	}
	clientB := &dummyDelayedClient{
		id:      "ClientB",
		waiting: 1,
		delay:   101.0,
	}
	clientC := &dummyDelayedClient{
		id:      "ClientC",
		waiting: 0,
		delay:   1.0,
	}
	clientD := &dummyDelayedClient{
		id:      "ClientD",
		waiting: 0,
		delay:   1.0,
	}

	cs := New([]transporter.Client{
		clientA, clientB, clientC, clientD,
	})

	cList := []Client{}

	cs.(*clients).iterate(0, cs.(*clients).pole.Len, func(c *client) bool {
		cList = append(cList, c)

		return true
	})

	for _, cl := range cList {
		cl.Request(
			func(transporter.HandlerConfig) transporter.Handler {
				return nil
			},
			transporter.RequestOption{
				Delay: func(addr string, delay float64, waiting uint64) {},
			},
		)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cs.Export(func(c Client) {
			clientsList = append(clientsList, c)
		})

		clientsList = clientsList[:0]
	}
}
