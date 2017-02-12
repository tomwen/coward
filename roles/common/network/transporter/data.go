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

package transporter

import (
	"io"

	"github.com/nickrio/coward/roles/common/network/buffer"
)

// Signal a type of channel that will be used to send signals
type Signal chan error

// RequestOption is the configuration for client request
type RequestOption struct {
	Buffer    buffer.Slice
	Canceller Signal
	Delay     func(avgConnectDelay float64, waiting uint64)
	Error     func(
		wantToRetry bool,
		wantToResetTransportConn bool,
		executeErr error,
	) (retry bool, reset bool, err error)
}

// ServeOption is the configuration for server handler
type ServeOption struct {
	Buffer       buffer.Slice
	Handler      HandlerBuilder
	Connected    func(clientInfo ServerClientInfo)
	Disconnected func(clientInfo ServerClientInfo, err error)
	Error        func(error) error
}

// ServeOptionBuilder will build a new ServeOption
type ServeOptionBuilder func(clientInfo ServerClientInfo) ServeOption

// ClientInfo contain methods which will return the meta information
// of a client
type ClientInfo interface {
	Name() string
}

// ClientConn represents a connection from client to server
type ClientConn interface {
	io.ReadWriteCloser
	ClientInfo

	Dial() error
	Rewind()
	Connected() bool
}

// ClientConnBuilder is a function that builds a new Client
type ClientConnBuilder func() ClientConn

// ServerConnAccepterMeta contains methods which returns meta info
// of a ServerConnAccepter
type ServerConnAccepterMeta interface {
	Name() string
}

// ServerConnAccepter represents a connection accepter that will
// accept connections and hand them over to Transporter server for
// later proccess
type ServerConnAccepter interface {
	ServerConnAccepterMeta

	Accept() (ServerClientConn, error)
	Close() error
}

// ServerConnListener represents a certain listener that will listen
// under control of a Transporter server
type ServerConnListener interface {
	Listen() (ServerConnAccepter, error)
}

// ServerClientInfo contains methods which will return information of
// a ServerClientConn
type ServerClientInfo interface {
	Name() string
}

// ServerClientConn represents a connection that initialized by a
// transporter connector which connects current Server
type ServerClientConn interface {
	ServerClientInfo
	io.ReadWriteCloser
}
