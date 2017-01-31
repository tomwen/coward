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

package listener

import (
	"net"
	"sync"
	"testing"
	"time"

	"github.com/nickrio/coward/common/logger"
	"github.com/nickrio/coward/roles/channel/common"
	"github.com/nickrio/coward/roles/common/network/transporter"
)

func TestUDPListen(t *testing.T) {
	wait := sync.WaitGroup{}
	transport := transporter.NewClient(
		"localhost", 5122, 100*time.Second, true,
		100, 3, 5*time.Second, nil, nil)

	u, uErr := NewUDP(common.ServerConfig{
		ID:          0,
		Interface:   net.ParseIP("127.0.0.1"),
		Port:        19191,
		Timeout:     60 * time.Second,
		Concurrence: 1,
		DefaultProc: nil,
		Transporter: transport,
		Logger:      logger.NewDitch(),
	})

	if uErr != nil {
		t.Error("UDP", uErr)

		return
	}

	go func() {
		time.Sleep(1 * time.Second)

		u.Close(&wait)
	}()

	u.Serve(&wait)
}
