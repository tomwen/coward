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
	"strconv"
	"time"

	"github.com/nickrio/coward/roles/common/network"
	"github.com/nickrio/coward/roles/common/network/communicator/common"
	"github.com/nickrio/coward/roles/common/network/transporter"
)

// serverConfig is configuration data for server
type serverConfig struct {
	ListenAddr     net.IP
	ListenPort     uint16
	ConnectTimeout time.Duration
	IdleTimeout    time.Duration
	Wrapper        common.ConnWrapper
	Disrupter      common.ConnDisrupter
}

// server is a Transporter server
type server struct {
	config *serverConfig
}

// NewServer returns a new Server Listener builder
func NewServer(
	listenAddr net.IP,
	listenPort uint16,
	connectTimeout time.Duration,
	idleTimeout time.Duration,
	wrapper common.ConnWrapper,
	disrupter common.ConnDisrupter,
) transporter.ServerConnListener {
	config := &serverConfig{
		ListenAddr:     listenAddr,
		ListenPort:     listenPort,
		ConnectTimeout: connectTimeout,
		IdleTimeout:    idleTimeout,
		Wrapper:        wrapper,
		Disrupter:      disrupter,
	}

	return &server{
		config: config,
	}
}

// Listen start listen on defined port and return a connection
// ServerConnAccepter for accepting incoming connections
func (s *server) Listen() (transporter.ServerConnAccepter, error) {
	listenConn, listenErr := net.Listen("tcp", net.JoinHostPort(
		s.config.ListenAddr.String(),
		strconv.FormatUint(uint64(s.config.ListenPort), 10)))

	if listenErr != nil {
		return nil, listenErr
	}

	return &serverAccepter{
		Config:      s.config,
		ListenConn:  listenConn,
		Connections: network.NewConnections(256),
	}, nil
}
