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

package address

import "net"

// UDP is encodable UDPAddr
type UDP net.UDPAddr

// Encode encode UDPAddr data into bytes
func (t *UDP) Encode(b []byte) (int, error) {
	ip4 := t.IP.To4()

	if ip4 != nil {
		return Default.Pack(IPv4, ip4, uint16(t.Port), b)
	}

	ip6 := t.IP.To16()

	if ip6 != nil {
		return Default.Pack(IPv6, ip6, uint16(t.Port), b)
	}

	return 0, nil
}

// Decode decodes a byte to build a UDPAddr
func (t *UDP) Decode(data []byte) (int, error) {
	atype, addr, port, offset, err := Default.Unpack(data)

	if err != nil {
		return 0, err
	}

	switch atype {
	case IPv4:
		fallthrough
	case IPv6:
		t.IP = addr
		t.Port = int(port)

	default:
		return 0, ErrUnsupportedAddress
	}

	return offset, err
}
