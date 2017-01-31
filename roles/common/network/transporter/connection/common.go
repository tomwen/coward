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
	"io"
	"net"
	"time"

	"github.com/nickrio/coward/roles/common/network/conn"
	"github.com/nickrio/coward/roles/common/network/transporter/common"
)

// wrap wraps input CONN
func wrap(
	raw net.Conn,
	wrapper common.ConnWrapper,
	disrupter common.ConnDisrupter,
	timeout time.Duration,
) (net.Conn, error) {
	// Wrapper layer order
	// EOF goes first so we can make sure there will be no conn after EOF
	// Then the timeout because now we can get correct error for timeout
	// Then the Error so we can filter some error out correctly for
	// transporter (Notice EOF and timed all deal and returning raw conn
	// errors, if we convert errors too soon, some of the error may be
	// overwritten back by them. If that happens, transporter will have no
	// idea of that's the error is actually happening)
	// Then Disrupter and then Encodec become last

	timeWrapped := conn.NewTimed(conn.NewEOF(raw))

	timeWrapped.SetTimeout(timeout)

	wrappedConn, wrappErr := wrapper(
		disrupter(conn.NewError(timeWrapped)))

	if wrappErr != nil {
		raw.Close()

		return nil, wrappErr
	}

	return wrappedConn, nil
}

// common is the connection rules that will be applied to
// both server to client and client to server connection
type base struct {
	net.Conn
}

// Read reads data from CONN and convert io.EOF and
// timeout error to the error which can be use by
// Transporter
func (s *base) Read(p []byte) (int, error) {
	rLen, rErr := s.Conn.Read(p)

	if rErr == nil {
		return rLen, rErr
	}

	switch rErr {
	case conn.ErrTimeout:
		fallthrough
	case conn.ErrReadTimeout:
		return rLen, ErrReadTimeout

	case conn.ErrUnconnectable:
		return rLen, ErrUnconnectable

	case conn.ErrRefused:
		return rLen, ErrRefused

	case conn.ErrResetted:
		return rLen, ErrResetted

	case conn.ErrAborted:
		return rLen, ErrAborted

	case io.EOF:
		return rLen, ErrEOF
	}

	return rLen, rErr
}

// Write writes data to CONN and convert io.EOF and
// timeout error to the error which can be use by
// Transporter
func (s *base) Write(p []byte) (int, error) {
	wLen, wErr := s.Conn.Write(p)

	if wErr == nil {
		return wLen, wErr
	}

	switch wErr {
	case conn.ErrTimeout:
		fallthrough
	case conn.ErrWriteTimeout:
		return wLen, ErrWriteTimeout

	case conn.ErrUnconnectable:
		return wLen, ErrUnconnectable

	case conn.ErrRefused:
		return wLen, ErrRefused

	case conn.ErrResetted:
		return wLen, ErrResetted

	case conn.ErrAborted:
		return wLen, ErrAborted

	case io.EOF:
		return wLen, ErrEOF
	}

	return wLen, wErr
}
