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

package common

import (
	"errors"
	"net"
)

var (
	// ErrUnsupportedCommand is throwed when the command is not
	// supported
	ErrUnsupportedCommand = errors.New(
		"Unsupported command")
)

// CTYPE is command type
type CTYPE byte

// CTYPES is how many CTYPE there is
const CTYPES = 5

// Consts used by command
const (
	Connect CTYPE = 0x01
	Bind    CTYPE = 0x02
	UDP     CTYPE = 0x03
)

// Commander is the command runer
type Commander func(
	addType ATYPE,
	addr []byte,
	port []byte,
	target []byte,
	rw net.Conn,
) error

// Command is the command hub
type Command struct {
	commands [CTYPES]Commander
}

// Register registers the command
func (c *Command) Register(cmd CTYPE, h Commander) {
	c.commands[cmd] = h
}

// Run run the command
func (c *Command) Run(
	cmd CTYPE,
	addrType ATYPE,
	addr []byte,
	port []byte,
	target []byte,
	rw net.Conn,
) error {
	if cmd >= CTYPES {
		return ErrUnsupportedCommand
	}

	if c.commands[cmd] == nil {
		return ErrUnsupportedCommand
	}

	return c.commands[cmd](addrType, addr, port, target, rw)
}
