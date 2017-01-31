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

package udp

import (
	"errors"
	"net"

	"github.com/nickrio/coward/roles/common/network/conn"
)

// Dispatcher errors
var (
	ErrDispatcherClosed = errors.New(
		"UDP Dispatcher has been closed")
)

const (
	maxPacketBufferSize = 4096
)

type udpClient interface {
	conn.UDPReadWriteCloser

	Send(readData) error
}

type rawRead struct {
	Addr *net.UDPAddr
	Len  int
	Buf  [maxPacketBufferSize]byte
}

type writeResult struct {
	Len int
	Err error
}

type writeData struct {
	Addr   *net.UDPAddr
	Data   []byte
	Result chan<- writeResult
}

type readData struct {
	Len  int
	Addr *net.UDPAddr
	Data []byte
}
