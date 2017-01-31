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

package channel

import (
	"net"
	"testing"
	"time"

	"github.com/nickrio/coward/common/logger"
	"github.com/nickrio/coward/roles/common/network"
	"github.com/nickrio/coward/roles/common/network/transporter"
	"github.com/nickrio/coward/roles/common/network/wrapper"
)

func TestChannelCreation(t *testing.T) {
	closeNotify := make(chan bool, 1)
	transport := transporter.NewClient(
		"localhost",
		9901,
		300*time.Second,
		true,
		64,
		3,
		5*time.Second,
		wrapper.PlainWrapper(nil),
		wrapper.ChaoticWrapper(nil))

	channel := New(transport, Config{
		DefaultTimeout: 300 * time.Second,
		Interface:      net.ParseIP("0.0.0.0"),
		Logger:         logger.NewDitch(),
		Channels: []Channel{
			Channel{
				ID:          0,
				Port:        1913,
				Protocol:    network.TCP,
				Timeout:     60 * time.Second,
				Concurrence: 1,
			},
			Channel{
				ID:          2,
				Port:        13915,
				Protocol:    network.UDP,
				Timeout:     60 * time.Second,
				Concurrence: 1,
			},
		},
	})

	spawnErr := channel.Spawn(closeNotify)

	if spawnErr != nil {
		t.Error("Can't spawn Channel due to error:", spawnErr)

		return
	}

	go func() {
		time.Sleep(5 * time.Second)

		channel.Unspawn()
	}()

	<-closeNotify
}
