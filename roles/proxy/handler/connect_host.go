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
	"strings"

	"github.com/nickrio/coward/common/types"
)

// Connect Host errors
var (
	ErrHostNotFound = errors.New(
		"Can't resolve the host address")

	ErrInvalidHostAddressPortLength = errors.New(
		"Invalid host address length")
)

func (h *handler) connectHost(
	buffer []byte,
	client io.ReadWriter,
	size uint16,
) error {
	// Connect Host format:
	//
	// +-----+------+------------+--------+
	// | CMD | SIZE |    Host    |  Port  |
	// +-----+------+------------+--------+
	// |  1  |  2   |    Host    |   2    |
	// +-----+------+------------+--------+
	//
	// Expected:
	// CMD:    4
	// SIZE:   Host + 2
	// Host:   String
	// Port:   uint16 (TCP Port)

	if size <= 2 {
		return ErrInvalidHostAddressPortLength
	}

	_, readAddrErr := io.ReadFull(client, buffer[:size])

	if readAddrErr != nil {
		return readAddrErr
	}

	address, addrEesloveErr := net.LookupIP(strings.TrimSpace(
		strings.ToLower(string(buffer[:size-2]))))

	if addrEesloveErr != nil {
		return ErrHostNotFound
	}

	port := types.EncodableUint16(0)

	decodeErr := port.DecodeBytes(buffer[size-2 : size])

	if decodeErr != nil {
		return ErrDecodingPortBytes
	}

	return h.connect(address[0], uint16(port), buffer, client)
}
