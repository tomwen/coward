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

package conn

import (
	"net"
	"time"

	"github.com/nickrio/coward/common/locked"
)

// Client is the connection from socks5 server to it's client
type Client interface {
	net.Conn
}

// ClientConfig is the config for client
type ClientConfig struct {
	Timeout time.Duration
	OnClose func()
}

// client implements Client
type client struct {
	net.Conn

	config ClientConfig
	closed locked.Boolean
}

// WrapClientConn wraps a raw net.Conn and turn it to client
// conn
func WrapClientConn(raw net.Conn, config ClientConfig) Client {
	timedConn := NewTimed(NewEOF(raw))

	timedConn.SetTimeout(config.Timeout)

	return &client{
		Conn:   NewError(timedConn),
		config: config,
		closed: locked.NewBool(false),
	}
}

// Close closes the connection to the client
func (c *client) Close() error {
	if !c.closed.Get() {
		c.closed.Set(true)

		c.config.OnClose()
	}

	return c.Conn.Close()
}
