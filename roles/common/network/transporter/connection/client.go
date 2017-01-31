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

package connection

import (
	"net"
	"time"

	"github.com/nickrio/coward/roles/common/network/transporter/common"
)

var (
	// ErrNotEstablished is throwed when client is not connected
	ErrNotEstablished = common.Error(
		"Connection not established")
)

// Client is the connection from client to server
type Client interface {
	net.Conn

	Connect(common.Address, common.ConnWrapper, common.ConnDisrupter) error
	Connected() bool
	Rewind()
}

// client implments Client
type client struct {
	base

	idleTimeout    time.Duration
	connectTimeout time.Duration
	connected      bool
	currentlyUsed  bool
	inStress       bool
	lastUse        time.Time
}

// NewClientConn wraps a client to server connection
func NewClientConn(
	idleTimeout time.Duration,
	connectTimeout time.Duration,
) Client {
	return &client{
		idleTimeout:    idleTimeout,
		connectTimeout: connectTimeout,
		connected:      false,
		currentlyUsed:  false,
		inStress:       false,
		lastUse:        time.Time{},
	}
}

// Connected return the value of connected mark
func (c *client) Connected() bool {
	if !c.lastUse.Add(c.idleTimeout).After(time.Now()) {
		c.connected = false

		return false
	}

	return c.connected
}

// Rewind rewind some setting for a new request
func (c *client) Rewind() {
	c.currentlyUsed = false
	c.inStress = true
}

// Connect connects the target server
func (c *client) Connect(
	addr common.Address,
	wrapper common.ConnWrapper,
	disrupter common.ConnDisrupter,
) error {
	addrStr, addrErr := addr.Get()

	if addrErr != nil {
		return addrErr
	}

	conn, connErr := net.DialTimeout("tcp", addrStr, c.connectTimeout)

	if connErr != nil {
		addr.TryRenew()

		return ErrUnconnectable
	}

	wrappedConn, wrapErr := wrap(conn, wrapper, disrupter, c.idleTimeout)

	if wrapErr != nil {
		return wrapErr
	}

	// Close the old connection handler
	if c.Conn != nil {
		c.Close()
	}

	c.Conn = wrappedConn
	c.connected = true

	return nil
}

// Read reads data from CONN and set connected mark
// according to error
func (c *client) Read(b []byte) (int, error) {
	if !c.connected {
		return 0, ErrNotEstablished
	}

	// Check if current connection is in stress. If it is, the
	// read timeout will be set to connect timeout so we can
	// fastly determine whether the connection is still ative or
	// not.
	// inStress mode should only be enabled when it's the first
	// time a new request started to use this current connection
	// and must be disabled when the client had succeed operation
	// with current connection
	if c.inStress {
		c.SetReadDeadline(time.Now().Add(c.connectTimeout))
	}

	rLen, rErr := c.base.Read(b)

	if rErr == nil {
		c.currentlyUsed = true
		c.inStress = false
		c.lastUse = time.Now()

		return rLen, rErr
	}

	if !c.currentlyUsed {
		c.Close()

		return rLen, ErrBroken
	}

	switch rErr {
	case ErrTimeout:
		fallthrough
	case ErrReadTimeout:
		fallthrough
	case ErrUnconnectable:
		fallthrough
	case ErrRefused:
		fallthrough
	case ErrResetted:
		fallthrough
	case ErrAborted:
		fallthrough
	case ErrEOF:
		c.Close()
	}

	return rLen, rErr
}

// Write writes data to CONN and set connected mark
// according to error
func (c *client) Write(b []byte) (int, error) {
	if !c.connected {
		return 0, ErrNotEstablished
	}

	wLen, wErr := c.base.Write(b)

	if wErr == nil {
		c.currentlyUsed = true
		c.lastUse = time.Now()

		return wLen, wErr
	}

	if !c.currentlyUsed {
		c.Close()

		return wLen, ErrBroken
	}

	switch wErr {
	case ErrTimeout:
		fallthrough
	case ErrWriteTimeout:
		fallthrough
	case ErrUnconnectable:
		fallthrough
	case ErrRefused:
		fallthrough
	case ErrResetted:
		fallthrough
	case ErrAborted:
		fallthrough
	case ErrEOF:
		c.Close()
	}

	return wLen, wErr
}

// Close closes current connection and set connected mark
// to false
func (c *client) Close() error {
	if !c.connected {
		return nil
	}

	c.connected = false

	return c.base.Close()
}
