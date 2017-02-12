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

package request

import (
	"fmt"
	"net"
	"time"

	ccommon "github.com/nickrio/coward/common"
	"github.com/nickrio/coward/roles/common/network/messaging"
	"github.com/nickrio/coward/roles/common/network/relay"
	"github.com/nickrio/coward/roles/common/network/transporter"
	"github.com/nickrio/coward/roles/socks5/common"
)

type connect struct {
	base

	command ccommon.Command
}

// NewConnectRequest creates a new connect request
func NewConnectRequest(
	config transporter.HandlerConfig,
	proc ccommon.Proccessors,
	client net.Conn,
	targetType common.ATYPE,
	targetAddr []byte,
	targetPort []byte,
	delayFeedback func(time.Duration),
) transporter.Handler {
	var cmdType ccommon.Command

	switch targetType {
	case common.IPv4:
		cmdType = messaging.ConnectIPv4

	case common.IPv6:
		cmdType = messaging.ConnectIPv6

	case common.Domain:
		cmdType = messaging.ConnectHost

	default:
		panic(fmt.Sprintf("Unknown target type: %v", targetType))
	}

	return &connect{
		base: base{
			buffer:        config.Buffer,
			proc:          proc,
			address:       append(targetAddr, targetPort...),
			server:        config.Server,
			client:        client,
			delayFeedback: delayFeedback,
			retryRequest:  false,
			resetTspConn:  false,
		},
		command: cmdType,
	}
}

func (c *connect) Handle() error {
	startTime := time.Now()

	// Enable retry and reset
	c.resetTspConn = true
	c.retryRequest = true

	// Send command
	_, writeErr := c.Write(
		c.server, c.command, c.address, c.buffer.Server.ExtendedBuffer)

	if writeErr != nil {
		return writeErr
	}

	// Read command result, should be succeed or failed
	dispErr := c.Dispatch(c.server, c.buffer.Server.Buffer, c.proc)

	if dispErr != nil {
		return dispErr
	}

	// Send delay feedback
	c.delayFeedback(time.Now().Sub(startTime))

	// Disable retry, becuase further request may generate side effect
	// Keep in mind the handler will be released when current request
	// is finished. So never mind those returns above. (Unless, we
	// changed this behave in the futher)
	c.retryRequest = false
	c.resetTspConn = false

	wErr := c.errorRespond(c.client, c.buffer.Client.ExtendedBuffer, 0)

	if wErr != nil {
		c.Write(c.server, messaging.EOF, nil, c.buffer.Server.ExtendedBuffer)

		return ErrFailedSendReadySignalToClient
	}

	return relay.NewTCPRelay(c.client, c.server, c.buffer, nil).Relay()
}
