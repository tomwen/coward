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
)

// Connections is a connection map container
type Connections interface {
	Serve()
	Close()
	Get(name string) (net.Conn, error)
	Put(name string, conn net.Conn) error
	Del(name string) (net.Conn, error)
	Iterate(iter func(name string, conn net.Conn))
}

type connectionsActionType byte

const (
	connectionActionGet  connectionsActionType = 0x01
	connectionActionPut  connectionsActionType = 0x02
	connectionActionDel  connectionsActionType = 0x03
	connectionActionIter connectionsActionType = 0x04
)

// Connections errors
var (
	ErrConnectionsConnectionNotFound = errors.New(
		"Connection not found")

	ErrConnectionsConnectionAlreadyExisted = errors.New(
		"Connection already existed")
)

type connectionsAction struct {
	Type   connectionsActionType
	Name   string
	Conn   net.Conn
	Func   func(string, net.Conn)
	Result chan connectionsActionResult
}

type connectionsActionResult struct {
	Conn net.Conn
	Err  error
}

type connections struct {
	connMap map[string]net.Conn
	reqChan chan connectionsAction
}

// NewConnections creates a new Connections
func NewConnections() Connections {
	return &connections{
		connMap: make(map[string]net.Conn, 256),
		reqChan: make(chan connectionsAction),
	}
}

// Serve start handle Connections request
func (c *connections) Serve() {
	for {
		action, reqOK := <-c.reqChan

		if !reqOK {
			return
		}

		switch action.Type {
		case connectionActionGet:
			conn, found := c.connMap[action.Name]

			if !found {
				action.Result <- connectionsActionResult{
					Conn: nil,
					Err:  ErrConnectionsConnectionNotFound,
				}
			} else {
				action.Result <- connectionsActionResult{
					Conn: conn,
					Err:  nil,
				}
			}

		case connectionActionPut:
			_, found := c.connMap[action.Name]

			if found {
				action.Result <- connectionsActionResult{
					Conn: nil,
					Err:  ErrConnectionsConnectionAlreadyExisted,
				}
			} else {
				c.connMap[action.Name] = action.Conn

				action.Result <- connectionsActionResult{
					Conn: action.Conn,
					Err:  nil,
				}
			}

		case connectionActionDel:
			conn, found := c.connMap[action.Name]

			if !found {
				action.Result <- connectionsActionResult{
					Conn: nil,
					Err:  ErrConnectionsConnectionNotFound,
				}
			} else {
				delete(c.connMap, action.Name)

				action.Result <- connectionsActionResult{
					Conn: conn,
					Err:  nil,
				}
			}

		case connectionActionIter:
			for key, conn := range c.connMap {
				action.Func(key, conn)
			}

			action.Result <- connectionsActionResult{
				Conn: nil,
				Err:  nil,
			}
		}
	}
}

// Close destory current Connections container
func (c *connections) Close() {
	close(c.reqChan)
}

// Get gets a conn item from container
func (c *connections) Get(name string) (net.Conn, error) {
	action := connectionsAction{
		Type:   connectionActionGet,
		Name:   name,
		Conn:   nil,
		Func:   nil,
		Result: make(chan connectionsActionResult),
	}

	c.reqChan <- action

	result := <-action.Result

	return result.Conn, result.Err
}

// Put puts a conn item to container
func (c *connections) Put(name string, conn net.Conn) error {
	action := connectionsAction{
		Type:   connectionActionPut,
		Name:   name,
		Conn:   conn,
		Func:   nil,
		Result: make(chan connectionsActionResult),
	}

	c.reqChan <- action

	result := <-action.Result

	return result.Err
}

// Del deletes a conn item from container
func (c *connections) Del(name string) (net.Conn, error) {
	action := connectionsAction{
		Type:   connectionActionDel,
		Name:   name,
		Conn:   nil,
		Func:   nil,
		Result: make(chan connectionsActionResult),
	}

	c.reqChan <- action

	result := <-action.Result

	return result.Conn, result.Err
}

// Iterate loop through all conn items in the container
func (c *connections) Iterate(iter func(name string, conn net.Conn)) {
	action := connectionsAction{
		Type:   connectionActionIter,
		Name:   "",
		Conn:   nil,
		Func:   iter,
		Result: make(chan connectionsActionResult),
	}

	c.reqChan <- action

	<-action.Result
}
