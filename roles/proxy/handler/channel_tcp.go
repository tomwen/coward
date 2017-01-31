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

package handler

import (
	"errors"
	"io"
	"net"

	"github.com/nickrio/coward/roles/common/network"
	"github.com/nickrio/coward/roles/common/network/messaging"
	"github.com/nickrio/coward/roles/common/network/relay"
)

// Channel errors
var (
	ErrInvalidChannelID = errors.New(
		"Invalid Channel ID")

	ErrRequestingUndefindedChannel = errors.New(
		"Requsting an undefined Channel")
)

func (h *handler) channelTCP(
	buffer []byte,
	client io.ReadWriter,
	size uint16,
) error {
	// Channel TCP format:
	//
	// +------------+
	// | Channel ID |
	// +------------+
	// |     1      |
	// +------------+
	//
	// Expected:
	// Channel ID:   1

	if size <= 0 {
		return ErrInvalidChannelID
	}

	// Read data for Channel ID
	_, readAddrErr := io.ReadFull(client, buffer[:size])

	if readAddrErr != nil {
		return readAddrErr
	}

	// Yeah, only read the first byte as Channel ID
	ch, chErr := h.channels.Get(buffer[0])

	if chErr != nil || ch.Protocol != network.TCP {
		return ErrRequestingUndefindedChannel
	}

	// Try to connect to the remote host
	remoteConn, dialErr := net.DialTimeout(
		"tcp", ch.Address, h.connectTimeout)

	if dialErr != nil {
		return ErrDestinationUnconnectable
	}

	defer remoteConn.Close()

	// Tell client we are set
	_, wErr := h.Write(client, messaging.OK, nil,
		h.buffer.Client.ExtendedBuffer)

	if wErr != nil {
		return ErrFailedSendConnectConfirmSignal
	}

	// Start relay
	return relay.NewTCPRelay(
		remoteConn, h.client, h.buffer, h.closeChan).Relay()
}
