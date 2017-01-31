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

	"github.com/nickrio/coward/common/types"
)

// Connect IPv4 errors
var (
	ErrInvalidIPv4AddrPortLength = errors.New(
		"Invalid IPv4 connect address length")
)

func (h *handler) connectIPv4(
	buffer []byte,
	client io.ReadWriter,
	size uint16,
) error {
	// Connect IPv4 format:
	//
	// +-----+------+----------+--------+
	// | CMD | SIZE |    IP    |  Port  |
	// +-----+------+----------+--------+
	// |  1  |  2   |    4     |   2    |
	// +-----+------+----------+--------+
	//
	// Expected:
	// CMD:    2
	// SIZE:   6
	// IP:     4 bytes
	// Port:   uint16 (TCP Port)

	if size < 6 {
		return ErrInvalidIPv4AddrPortLength
	}

	_, readAddrErr := io.ReadFull(client, buffer[:size])

	if readAddrErr != nil {
		return readAddrErr
	}

	ipv4Addr := net.IPv4(
		buffer[0],
		buffer[1],
		buffer[2],
		buffer[3],
	)

	port := types.EncodableUint16(0)

	decodeErr := port.DecodeBytes(buffer[4:6])

	if decodeErr != nil {
		return ErrDecodingPortBytes
	}

	return h.connect(ipv4Addr, uint16(port), buffer, client)
}
