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
	"strconv"

	"github.com/nickrio/coward/roles/common/network/conn"
	"github.com/nickrio/coward/roles/common/network/messaging"
	"github.com/nickrio/coward/roles/common/network/relay"
)

// Connect errors
var (
	ErrInvalidAddress = errors.New(
		"Invalid IP address")

	ErrDecodingPortBytes = errors.New(
		"Failed to decode the port bytes")

	ErrLoopbackAddressIsForbidden = errors.New(
		"Access loopback address is forbidden")

	ErrZeroAddressIsForbidden = errors.New(
		"Access zero address is forbidden")

	ErrZeroPortIsForbidden = errors.New(
		"Access zero port is forbidden")

	ErrDestinationUnconnectable = errors.New(
		"Can't connect to destination")

	ErrFailedSendConnectConfirmSignal = errors.New(
		"Can't send connect confirm signal to client")

	ErrRemoteConnectionClosed = errors.New(
		"Remote connection is closed")
)

func (h *handler) connect(
	ip net.IP,
	port uint16,
	buffer []byte,
	client io.ReadWriter,
) error {
	if ip == nil {
		return ErrInvalidAddress
	}

	if ip.IsLoopback() {
		return ErrLoopbackAddressIsForbidden
	}

	if ip.IsUnspecified() {
		return ErrLoopbackAddressIsForbidden
	}

	if port <= 0 {
		return ErrZeroPortIsForbidden
	}

	target, targetConnErr := net.DialTimeout("tcp", net.JoinHostPort(
		ip.String(), strconv.FormatUint(uint64(port), 10)), h.connectTimeout)

	if targetConnErr != nil {
		return ErrDestinationUnconnectable
	}

	targetConn := conn.NewTimed(conn.NewError(target))

	targetConn.SetTimeout(h.idleTimeout)

	// Drop the connection if client aborted or disconnected
	defer targetConn.Close()

	_, wErr := h.Write(client, messaging.OK, nil,
		h.buffer.Client.ExtendedBuffer)

	if wErr != nil {
		return ErrFailedSendConnectConfirmSignal
	}

	return relay.NewTCPRelay(
		targetConn, h.client, h.buffer, h.closeChan).Relay()
}
