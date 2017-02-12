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
	"syscall"
)

// Connection errors
var (
	ErrTimeout = newErrorConnError(
		"Connection timed out")

	ErrReadTimeout = newErrorConnError(
		"Connection read timed out")

	ErrWriteTimeout = newErrorConnError(
		"Connection write timed out")

	ErrUnconnectable = newErrorConnError(
		"Unable to connect to the host")

	ErrRefused = newErrorConnError(
		"Connection refused")

	ErrResetted = newErrorConnError(
		"Connection resetted")

	ErrAborted = newErrorConnError(
		"Connection aborted")
)

type errorConnOpType byte

// OPtypes
const (
	errorConnOpOther errorConnOpType = 0
	errorConnOpRead  errorConnOpType = 1
	errorConnOpWrite errorConnOpType = 2
)

type errorConn struct {
	net.Conn
}

// ErrorConnError is errors for ErrorConn
type ErrorConnError interface {
	error

	IsErrorConnError() bool
}

type errorConnError struct {
	message string
}

func (e *errorConnError) Error() string {
	return e.message
}

func (e *errorConnError) IsErrorConnError() bool {
	return true
}

func newErrorConnError(message string) ErrorConnError {
	return &errorConnError{
		message: message,
	}
}

// NewError creates a new error CONN
func NewError(c net.Conn) net.Conn {
	return &errorConn{
		Conn: c,
	}
}

// convertErr converts those error returned from lower call to
// the error we can understand
func (e *errorConn) convertErr(err error, op errorConnOpType) error {
	if netErr, netErrOK := err.(net.Error); netErrOK && netErr.Timeout() {
		if op == errorConnOpRead {
			return ErrReadTimeout
		} else if op == errorConnOpWrite {
			return ErrWriteTimeout
		}

		return ErrTimeout
	}

	switch eType := err.(type) {
	case *net.OpError:
		if eType.Op == "dial" {
			return ErrUnconnectable
		}

		return ErrRefused

	case syscall.Errno:
		switch eType {
		case syscall.ECONNREFUSED:
			return ErrRefused

		case syscall.ECONNRESET:
			return ErrResetted

		case syscall.ECONNABORTED:
			return ErrAborted
		}
	}

	return err
}

// Read reads data from conn, and convert net error to better
// type
func (e *errorConn) Read(p []byte) (int, error) {
	rLen, rErr := e.Conn.Read(p)

	if rErr == nil {
		return rLen, rErr
	}

	return rLen, e.convertErr(rErr, errorConnOpRead)
}

// Write writes data to conn, and convert net error to better
// type
func (e *errorConn) Write(p []byte) (int, error) {
	wLen, wErr := e.Conn.Write(p)

	if wErr == nil {
		return wLen, wErr
	}

	return wLen, e.convertErr(wErr, errorConnOpWrite)
}
