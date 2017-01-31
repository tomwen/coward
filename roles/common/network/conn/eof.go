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
	"io"
	"net"
)

type eof struct {
	net.Conn

	eof bool
}

// NewEOF creates a new EOF CONN
func NewEOF(c net.Conn) net.Conn {
	return &eof{
		Conn: c,
		eof:  false,
	}
}

// Read read from CONN
func (e *eof) Read(p []byte) (int, error) {
	if e.eof {
		return 0, io.EOF
	}

	rLen, rErr := e.Conn.Read(p)

	if rErr == nil {
		return rLen, rErr
	}

	if rErr == io.EOF {
		e.eof = true
	}

	return rLen, rErr
}

// Write writes from CONN
func (e *eof) Write(p []byte) (int, error) {
	if e.eof {
		return 0, io.EOF
	}

	wLen, wErr := e.Conn.Write(p)

	if wErr == nil {
		return wLen, wErr
	}

	if wErr == io.EOF {
		e.eof = true
	}

	return wLen, wErr
}

// Close closes the CONN
func (e *eof) Close() error {
	if e.eof {
		return io.EOF
	}

	closeErr := e.Conn.Close()

	if closeErr != nil {
		return closeErr
	}

	e.eof = true

	return nil
}
