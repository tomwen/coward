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

package messaging

import "github.com/nickrio/coward/common"

// Available commons
const (
	NOP            common.Command = 0 // Actions
	ChannelTCP     common.Command = 1
	ChannelUDP     common.Command = 2
	ConnectIPv4    common.Command = 3
	ConnectIPv6    common.Command = 4
	ConnectHost    common.Command = 5
	RelayUDP       common.Command = 6
	ResolveHost    common.Command = 7
	PingIPv4       common.Command = 8
	PingIPv6       common.Command = 9
	PingHost       common.Command = 10
	Unleash        common.Command = 11
	Streaming      common.Command = 12 // Exchange type
	Datagram       common.Command = 13
	OK             common.Command = 14 // Status
	EOF            common.Command = 15
	Error          common.Command = 16
	Closed         common.Command = 17
	InternalError  common.Command = 18
	Forbidden      common.Command = 19
	Unconnectable  common.Command = 20
	Timeout        common.Command = 21
	Unsupported    common.Command = 22
	Invalid        common.Command = 23
	UnknownCommand common.Command = 24
)
