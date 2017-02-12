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

package tcp

import (
	"net"

	"github.com/nickrio/coward/roles/common/network"
	"github.com/nickrio/coward/roles/common/network/transporter"
)

// serverAccepter will accept incoming connect request,
// wrap it to get it ready and return it as a Transporter
// ServerClientConn
type serverAccepter struct {
	Config      *serverConfig
	ListenConn  net.Listener
	Connections network.Connections
}

// Name returns the name of current Accepter
func (s *serverAccepter) Name() string {
	return s.ListenConn.Addr().String()
}

// Accept accepts incoming connections
func (s *serverAccepter) Accept() (transporter.ServerClientConn, error) {
	aConn, aErr := s.ListenConn.Accept()

	if aErr != nil {
		return nil, aErr
	}

	wrapped, wrapErr := wrapConn(
		aConn,
		s.Config.Wrapper,
		s.Config.Disrupter,
		s.Config.IdleTimeout,
	)

	if wrapErr != nil {
		aConn.Close()

		return nil, wrapErr
	}

	clientAddr := wrapped.RemoteAddr().String()
	clientWrapped := &serverConn{
		Conn:          wrapped,
		RemoteAddress: clientAddr,
		OnClose: func(name string) {
			s.Connections.Del(name)
		},
	}

	s.Connections.Put(clientAddr, clientWrapped)

	return clientWrapped, nil
}

// Close closes current accepter. This will shutdown the listening
// connection and all file descripters associated with it
func (s *serverAccepter) Close() error {
	cErr := s.ListenConn.Close()

	if cErr != nil {
		return cErr
	}

	// Must call close in a routine or we will doomed to
	// dead lock
	s.Connections.Iterate(func(name string, conn net.Conn) {
		go conn.Close()
	})

	return nil
}
