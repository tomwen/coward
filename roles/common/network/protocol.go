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

package network

import (
	"errors"
	"strings"
)

// Protocol is the ... protocol. TCP and UDP basicly
type Protocol byte

// Protocol errors
var (
	ErrProtocolUndefined = errors.New(
		"Undefined protocol")
)

// Protocols
const (
	TCP Protocol = 0x01
	UDP Protocol = 0x02
)

// Some consts related to Protocol
const (
	MaxProtocolID = Protocol(8)
)

// String returns string type according to Protocol
func (p *Protocol) String() string {
	switch *p {
	case TCP:
		return "TCP"
	case UDP:
		return "UDP"
	}

	return "UNKNOWN"
}

// GetProtocolByString returns specific Protocol according to inputted string
func GetProtocolByString(pro string) (Protocol, error) {
	switch strings.ToLower(pro) {
	case "tcp":
		return TCP, nil
	case "udp":
		return UDP, nil
	}

	return 0, ErrProtocolUndefined
}
