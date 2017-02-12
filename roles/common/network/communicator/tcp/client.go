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
	"errors"
	"net"
	"time"

	"github.com/nickrio/coward/common/locked"
	"github.com/nickrio/coward/roles/common/network/communicator/common"
	"github.com/nickrio/coward/roles/common/network/transporter"
)

// Client errors
var (
	ErrClientNoConnection = errors.New(
		"Client no connection")

	ErrClientFailedToConnection = errors.New(
		"Client couldn't dial to the remote host")
)

// clientConfig is configuration data used by client
type clientConfig struct {
	dialer         common.Dialer
	connectTimeout time.Duration
	idleTimeout    time.Duration
	wrapper        common.ConnWrapper
	disrupter      common.ConnDisrupter
}

// client implements Transporter Client
type client struct {
	net.Conn

	config    *clientConfig
	used      locked.Boolean
	connected bool
}

// NewClientBuilder builds a new Client builder
func NewClientBuilder(
	host string,
	port uint16,
	connectTimeout time.Duration,
	idleTimeout time.Duration,
	wrapper common.ConnWrapper,
	disrupter common.ConnDisrupter,
) func() transporter.ClientConn {
	config := &clientConfig{
		dialer:         common.NewDialer(host, port),
		connectTimeout: connectTimeout,
		idleTimeout:    idleTimeout,
		wrapper:        wrapper,
		disrupter:      disrupter,
	}

	return func() transporter.ClientConn {
		return &client{
			Conn:      nil,
			config:    config,
			used:      locked.NewBool(false),
			connected: false,
		}
	}
}

// Dial connects to the Transporter server
func (c *client) Dial() error {
	if c.Conn != nil {
		// Close the previous connection blindly, as
		// we don't need to know the close result.
		// Everythig will be OK as long as the connection
		// is closed.
		c.Close()

		c.Conn = nil
	}

	newConn, dialErr := c.config.dialer.Dial(
		"tcp", c.config.connectTimeout)

	if dialErr != nil {
		return ErrClientFailedToConnection
	}

	wrappedConn, wrappedErr := wrapConn(
		newConn,
		c.config.wrapper,
		c.config.disrupter,
		c.config.idleTimeout,
	)

	if wrappedErr != nil {
		wrappedConn.Close()

		return wrappedErr
	}

	c.Conn = wrappedConn
	c.connected = true

	return nil
}

// Rewind resets the status of current Transporter without
// closing it
func (c *client) Rewind() {
	c.used.Set(false)
}

// Connected returns whether or not current client is connected
// with a Transporter server
func (c *client) Connected() bool {
	return c.connected
}

// Name returns ID of current client
func (c *client) Name() string {
	return c.Conn.LocalAddr().String()
}

// Read reads data from source, and close connection when any
// error happened
func (c *client) Read(b []byte) (int, error) {
	if !c.used.Get() {
		// Set one time timeout
		c.SetDeadline(time.Now().Add(c.config.connectTimeout))

		c.used.Set(true)
	}

	rLen, rErr := c.Conn.Read(b)

	if rErr != nil {
		c.Close()
	}

	return rLen, rErr
}

// Write writes data to the source, and close connection when
// any error happened
func (c *client) Write(b []byte) (int, error) {
	wLen, wErr := c.Conn.Write(b)

	if wErr != nil {
		c.Close()
	}

	return wLen, wErr
}

// Close closes current client connection
func (c *client) Close() error {
	if c.Conn == nil {
		return ErrClientNoConnection
	}

	// Don't care if Conn.Close actually works.
	// This is because the Conn.Close may fail due to a
	// already broken connection etc
	c.connected = false

	return c.Conn.Close()
}
