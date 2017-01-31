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
	"github.com/nickrio/coward/common"
	"github.com/nickrio/coward/roles/common/network/conn"
	"github.com/nickrio/coward/roles/common/network/messaging"
	"github.com/nickrio/coward/roles/common/network/relay"
	"github.com/nickrio/coward/roles/common/network/transporter"
)

type udp struct {
	base

	listener  conn.UDPReadWriteCloser
	closeChan chan bool
}

type udpHandler struct{}

func (u *udpHandler) Ready() error {
	return nil
}

func (u *udpHandler) Quitter() relay.SignalChan {
	return nil
}

// Send read data from internal source and send it to client
func (u *udpHandler) Send(
	c conn.UDPReadWriter, data []byte, writeBuf []byte) (int, error) {
	// Notice here, we trying to write data to a nil UDPAddr
	// This is fine because the dispatcher will automatically
	// fillin this setting whatever if it's correct
	return c.WriteToUDP(data, nil)
}

// Receive receives data from client and send it to the remote through
// tcp tunnel
func (u *udpHandler) Receive(
	c conn.UDPReadWriter, result []byte, readBuf []byte) (int, error) {
	// Notice we ignored the address is because the dispatcher will
	// deal with address related thing by itself, so we don't have to
	rLen, _, rErr := c.ReadFromUDP(readBuf)

	if rErr != nil {
		return 0, rErr
	}

	return copy(result[0:], readBuf[0:rLen]), nil
}

// NewUDPRequest creates a new UDP channel request
func NewUDPRequest(
	channelID byte,
	udpListener conn.UDPReadWriteCloser,
	proc common.Proccessors,
	closeChan chan bool,
) transporter.HandlerBuilder {
	return func(config transporter.HandlerConfig) transporter.Handler {
		return &udp{
			base: base{
				channelID:    channelID,
				buffer:       config.Buffer,
				proc:         proc,
				server:       config.Server,
				retryRequest: false,
				resetTspConn: false,
			},
			listener:  udpListener,
			closeChan: closeChan,
		}
	}
}

func (u *udp) Handle() error {
	u.resetTspConn = true
	u.retryRequest = true

	_, writeErr := u.Write(u.server, messaging.ChannelUDP,
		[]byte{u.channelID}, u.buffer.Server.ExtendedBuffer)

	if writeErr != nil {
		return writeErr
	}

	dispErr := u.Dispatch(u.server, u.buffer.Server.Buffer, u.proc)

	if dispErr != nil {
		return dispErr
	}

	u.resetTspConn = false
	u.retryRequest = false

	return relay.NewUDPRelay(
		&udpHandler{},
		u.listener,
		u.server,
		u.buffer,
		u.closeChan,
	).Relay()
}
