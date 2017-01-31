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
	"io"
	"net"
)

// UDP Client errors
var (
	ErrClientAlreadyClosed = errors.New(
		"UDP Client already been closed")

	ErrClientSendCancelled = errors.New(
		"UDP Client has cancelled the data receive")
)

// client is an artifical UDP connection
type client struct {
	targetAddr      *net.UDPAddr
	readChan        chan readData
	writeChan       chan<- writeData
	sendCancelChan  chan bool
	readCancelChan  chan bool
	writeResultChan chan writeResult
	closed          bool
}

// newClient creates a new an artifical UDP connection
func newClient(
	targetAddr *net.UDPAddr,
	wChan chan<- writeData,
) *client {
	return &client{
		targetAddr:      targetAddr,
		readChan:        make(chan readData),
		writeChan:       wChan,
		sendCancelChan:  make(chan bool),
		readCancelChan:  make(chan bool),
		writeResultChan: make(chan writeResult),
		closed:          false,
	}
}

// WriteToUDP writes data to UDP connection, the address field will be ignored
func (c *client) WriteToUDP(b []byte, addr *net.UDPAddr) (int, error) {
	if c.closed {
		return 0, io.EOF
	}

	c.writeChan <- writeData{
		Addr:   c.targetAddr,
		Data:   b,
		Result: c.writeResultChan,
	}

	result := <-c.writeResultChan

	return result.Len, result.Err
}

// ReadFromUDP reads data from UDP connection
func (c *client) ReadFromUDP(b []byte) (int, *net.UDPAddr, error) {
	if c.closed {
		return 0, nil, io.EOF
	}

	select {
	case rData, ok := <-c.readChan:
		if !ok {
			c.closed = true

			return 0, nil, io.EOF
		}

		copy(b, rData.Data[:rData.Len])

		return rData.Len, rData.Addr, nil

	case <-c.readCancelChan:
		c.closed = true

		return 0, nil, io.EOF
	}
}

// Send send data to current UDP connection, this method is thread-safe
func (c *client) Send(r readData) error {
	if c.closed {
		return ErrClientAlreadyClosed
	}

	select {
	case c.readChan <- r:
		return nil

	case _, ok := <-c.sendCancelChan:
		if !ok {
			c.closed = true

			close(c.readChan)
		}

		return ErrClientSendCancelled
	}
}

// Delete tells current UDP connection it has been removed
func (c *client) Delete() {
	c.closed = true

	select {
	case c.sendCancelChan <- true:
	default:
	}

	close(c.sendCancelChan)
}

// Close shuts down current connection
func (c *client) Close() error {
	if c.closed {
		return ErrClientAlreadyClosed
	}

	c.closed = true

	select {
	case c.readCancelChan <- true:
	default:
	}

	close(c.readCancelChan)

	return nil
}
