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
)

// Timed is a conn that will automatically apply connection timeout
type Timed interface {
	net.Conn
	SetTimeout(time.Duration)
}

// timed implements TimeoutConn
type timed struct {
	net.Conn

	timeout              time.Duration
	onetimeReadDeadline  time.Time
	onetimeWriteDeadline time.Time
}

// zeroTime is the zero value of the time type
var zeroTime = time.Time{}

// NewTimed creates a new Timed CONN
func NewTimed(raw net.Conn) Timed {
	return &timed{
		Conn:                 raw,
		timeout:              time.Duration(0),
		onetimeReadDeadline:  zeroTime,
		onetimeWriteDeadline: zeroTime,
	}
}

// SetTimeout set the timeout
func (t *timed) SetTimeout(timeout time.Duration) {
	t.timeout = timeout
}

// SetDeadline set both read and write dead line for one time
func (t *timed) SetDeadline(deadline time.Time) error {
	t.SetReadDeadline(deadline)
	t.SetWriteDeadline(deadline)

	return nil
}

// SetReadDeadline set read dead line for one time
func (t *timed) SetReadDeadline(deadline time.Time) error {
	t.onetimeReadDeadline = deadline

	return nil
}

// SetWriteDeadline set write dead line for one time
func (t *timed) SetWriteDeadline(deadline time.Time) error {
	t.onetimeWriteDeadline = deadline

	return nil
}

// Read read from conn
func (t *timed) Read(p []byte) (int, error) {
	var dlErr error

	if t.onetimeReadDeadline != zeroTime {
		dlErr = t.Conn.SetReadDeadline(t.onetimeReadDeadline)

		t.onetimeReadDeadline = zeroTime
	} else {
		dlErr = t.Conn.SetReadDeadline(time.Now().Add(t.timeout))
	}

	if dlErr != nil {
		return 0, dlErr
	}

	return t.Conn.Read(p)
}

// Write write to conn
func (t *timed) Write(p []byte) (int, error) {
	var dlErr error

	if t.onetimeWriteDeadline != zeroTime {
		dlErr = t.Conn.SetWriteDeadline(t.onetimeWriteDeadline)

		t.onetimeWriteDeadline = zeroTime
	} else {
		dlErr = t.Conn.SetWriteDeadline(time.Now().Add(t.timeout))
	}

	if dlErr != nil {
		return 0, dlErr
	}

	return t.Conn.Write(p)
}
