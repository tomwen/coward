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

// server is the server to client connection
// server CONN will automatically close current connection when
// following error happens:
//  ErrTimeout, ErrReadTimeout, ErrUnconnectable, ErrRefused,
//  ErrResetted, ErrAborted, ErrClosed
type server struct {
	base

	closed bool
}

// WrapServerConn wraps a server to client connection
// parameter c should be the data source (EncodedConn wrapped
// CONN in this case)
func WrapServerConn(
	raw net.Conn,
	wrapper common.ConnWrapper,
	disrupter common.ConnDisrupter,
	timeout time.Duration,
) (net.Conn, error) {
	wrappedConn, wrappErr := wrap(raw, wrapper, disrupter, timeout)

	if wrappErr != nil {
		return nil, wrappErr
	}

	return &server{
		base: base{
			Conn: wrappedConn,
		},
		closed: false,
	}, nil
}

// Read reads data from CONN and automatically close CONN when
// selected error happens
func (s *server) Read(b []byte) (int, error) {
	rLen, rErr := s.base.Read(b)

	if rErr == nil {
		return rLen, rErr
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
		s.Close()
	}

	return rLen, rErr
}

// Write write data to CONN and automatically close CONN when
// selected error happens
func (s *server) Write(b []byte) (int, error) {
	wLen, wErr := s.base.Write(b)

	if wErr == nil {
		return wLen, wErr
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
		s.Close()
	}

	return wLen, wErr
}

// Close closes current connection
func (s *server) Close() error {
	if s.closed {
		return ErrReclosing
	}

	closeErr := s.base.Close()

	if closeErr != nil {
		return closeErr
	}

	s.closed = true

	return nil
}
