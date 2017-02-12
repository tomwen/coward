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
	"net"
	"sync"
)

// Connections is a connection map container
type Connections interface {
	Get(name string) (net.Conn, error)
	Put(name string, conn net.Conn) error
	Del(name string) (net.Conn, error)
	Iterate(iter func(name string, conn net.Conn))
}

// Connections errors
var (
	ErrConnectionsConnectionNotFound = errors.New(
		"Connection not found")

	ErrConnectionsConnectionAlreadyExisted = errors.New(
		"Connection already existed")
)

type connections struct {
	connMap map[string]net.Conn
	lock    sync.RWMutex
}

// NewConnections creates a new Connections
func NewConnections(size uint) Connections {
	return &connections{
		connMap: make(map[string]net.Conn, size),
		lock:    sync.RWMutex{},
	}
}

// Get gets a conn item from container
func (c *connections) Get(name string) (net.Conn, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	conn, found := c.connMap[name]

	if !found {
		return nil, ErrConnectionsConnectionNotFound
	}

	return conn, nil
}

// Put puts a conn item to container
func (c *connections) Put(name string, conn net.Conn) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	_, found := c.connMap[name]

	if found {
		return ErrConnectionsConnectionAlreadyExisted
	}

	c.connMap[name] = conn

	return nil
}

// Del deletes a conn item from container
func (c *connections) Del(name string) (net.Conn, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	conn, found := c.connMap[name]

	if !found {
		return nil, ErrConnectionsConnectionNotFound
	}

	delete(c.connMap, name)

	return conn, nil
}

// Iterate loop through all conn items in the container
func (c *connections) Iterate(iter func(name string, conn net.Conn)) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	for k, v := range c.connMap {
		iter(k, v)
	}
}
