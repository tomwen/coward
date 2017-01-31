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
	"net"

	"github.com/nickrio/coward/common"
	"github.com/nickrio/coward/roles/common/network/messaging"
	"github.com/nickrio/coward/roles/common/network/relay"
	"github.com/nickrio/coward/roles/common/network/transporter"
)

type tcp struct {
	base

	client net.Conn
}

// NewTCPRequest creates a new TCP Channel request
func NewTCPRequest(
	channelID byte,
	client net.Conn,
	proc common.Proccessors,
) transporter.HandlerBuilder {
	return func(config transporter.HandlerConfig) transporter.Handler {
		return &tcp{
			base: base{
				channelID:    channelID,
				buffer:       config.Buffer,
				proc:         proc,
				server:       config.Server,
				retryRequest: false,
				resetTspConn: false,
			},
			client: client,
		}
	}
}

func (t *tcp) Handle() error {
	t.resetTspConn = true
	t.retryRequest = true

	// Ask server to open Channel connection for us
	_, writeErr := t.Write(t.server, messaging.ChannelTCP,
		[]byte{t.channelID}, t.buffer.Server.ExtendedBuffer)

	if writeErr != nil {
		return writeErr
	}

	// See what's server say
	dispErr := t.Dispatch(t.server, t.buffer.Server.Buffer, t.proc)

	if dispErr != nil {
		return dispErr
	}

	t.resetTspConn = false
	t.retryRequest = false

	// OK, start relay
	return relay.NewTCPRelay(t.client, t.server, t.buffer, nil).Relay()
}
