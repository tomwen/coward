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

// Connect IPv6 errors
var (
	ErrInvalidIPv6AddrPortLength = errors.New(
		"Invalid IPv6 connect address length")
)

func (h *handler) connectIPv6(
	buffer []byte,
	client io.ReadWriter,
	size uint16,
) error {
	// Connect IPv6 format:
	//
	// +-----+------+----------+--------+
	// | CMD | SIZE |    IP    |  Port  |
	// +-----+------+----------+--------+
	// |  1  |  2   |    16    |   2    |
	// +-----+------+----------+--------+
	//
	// Expected:
	// CMD:    2
	// SIZE:   18
	// IP:     16 bytes
	// Port:   uint16 (TCP Port)

	if size < 18 {
		return ErrInvalidIPv6AddrPortLength
	}

	_, readAddrErr := io.ReadFull(client, buffer[:size])

	if readAddrErr != nil {
		return readAddrErr
	}

	ipv6Addr := net.IP(buffer[:16])

	port := types.EncodableUint16(0)

	decodeErr := port.DecodeBytes(buffer[16:18])

	if decodeErr != nil {
		return ErrDecodingPortBytes
	}

	return h.connect(ipv6Addr, uint16(port), buffer, client)
}
