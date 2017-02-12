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

package transporter

// Error is transporter error
type Error interface {
	error

	Is(err error) bool
	IsTransporterError() bool
	Raw() error
}

// wrappedError implements Error
type wrappedError struct {
	err error
}

// underError implements Error too
type underError struct {
	err   Error
	under error
}

// errored also implements Error
type errored struct {
	msg string
}

// Error returns error message
func (w *wrappedError) Error() string {
	return "Transporter Error: " + w.err.Error()
}

// Is compare if inputted error is current error
func (w *wrappedError) Is(err error) bool {
	return w.err == err
}

// IsTransporterError is the Interface Characteristic
// for Transpoter Errors
func (w *wrappedError) IsTransporterError() bool {
	return true
}

// Raw returns the raw error which underlaying current
// Transpoter error
func (w *wrappedError) Raw() error {
	return w.err
}

// Error returns error message
func (u *underError) Error() string {
	return u.err.Error() + ": " + u.under.Error()
}

// Is compare if inputted error is current error
func (u *underError) Is(err error) bool {
	return u.err == err
}

// IsTransporterError is the Interface Characteristic
// for Transpoter Errors
func (u *underError) IsTransporterError() bool {
	return true
}

// Raw returns the raw error which underlaying current
// Transpoter error
func (u *underError) Raw() error {
	return u.under
}

// Error returns error message
func (e *errored) Error() string {
	return e.msg
}

// Is compare if inputted error is current error
func (e *errored) Is(err error) bool {
	return e == err
}

// IsTransporterError is the Interface Characteristic
// for Transpoter Errors
func (e *errored) IsTransporterError() bool {
	return true
}

// Raw returns the raw error which underlaying current
// Transpoter error
func (e *errored) Raw() error {
	return e
}

// WrapError wraps inputted error to make it a Transporter error
func WrapError(err error) Error {
	return &wrappedError{
		err: err,
	}
}

// UnderError wraps inputted error as a sub error of a new Transporter error
func UnderError(err Error, under error) Error {
	return &underError{
		err:   err,
		under: under,
	}
}

// NewError creates a new Transporter error
func NewError(message string) Error {
	return &errored{
		msg: message,
	}
}
