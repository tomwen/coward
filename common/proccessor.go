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
	"io"
)

var (
	// ErrCommandUnsupported is throwed when command is
	// not registered
	ErrCommandUnsupported = errors.New(
		"Unregistered command")
)

// Command is command type
type Command byte

// Proccessor is a function that will handle a specific Command
type Proccessor func(buffer []byte, rw io.ReadWriter, size uint16) error

// Proccessors is a hub of proccessors
type Proccessors interface {
	Register(cmd Command, handler Proccessor) Proccessors
	Execute(cmd Command, buffer []byte, rw io.ReadWriter, size uint16) error
}

const (
	maxCommands Command = 32
)

// proccessors implements Proccessors
type proccessors struct {
	commands [maxCommands]Proccessor
}

// NewProccessors creates a new Proccessors
func NewProccessors() Proccessors {
	return &proccessors{
		commands: [maxCommands]Proccessor{},
	}
}

// Register binds a Proccessor to a command
func (p *proccessors) Register(cmd Command, proc Proccessor) Proccessors {
	p.commands[cmd] = proc

	return p
}

// Execute run a Proccessor according to comand
func (p *proccessors) Execute(
	cmd Command,
	buffer []byte,
	rw io.ReadWriter,
	size uint16,
) error {
	if cmd >= maxCommands {
		return ErrCommandUnsupported
	}

	if p.commands[cmd] == nil {
		return ErrCommandUnsupported
	}

	return p.commands[cmd](buffer, rw, size)
}
