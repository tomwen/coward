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
	"io"

	"github.com/nickrio/coward/common"
	"github.com/nickrio/coward/roles/common/network/messaging"
)

// Default Proc errors
var (
	ErrProcRemoteTargetClosed = errors.New(
		"Remote destination is closed")

	ErrProcServerInternalError = errors.New(
		"Server has failed to proccess the request due to internal error")

	ErrProcServerRefused = errors.New(
		"Server refused the request")

	ErrProcRemoteTargetUnconnectable = errors.New(
		"Server has failed to connect to the remote destination")

	ErrProcTimeout = errors.New(
		"Server to target connection is timed out")

	ErrProcUnsupported = errors.New(
		"Unsupported request")

	ErrProcInvalid = errors.New(
		"Invalid request")

	ErrProcUnsupportedCommand = errors.New(
		"Server doesn't support that command")
)

// GetDefaultProc returns a default Proccessors Group for handling transporter
// traffic
func GetDefaultProc() common.Proccessors {
	return common.NewProccessors().
		Register(messaging.OK,
			func(buffer []byte, rw io.ReadWriter, size uint16) error {
				if size > 0 {
					io.ReadFull(rw, buffer[:size]) // Drain errors
				}

				return nil
			}).
		Register(messaging.EOF,
			func(buffer []byte, rw io.ReadWriter, size uint16) error {
				if size > 0 {
					io.ReadFull(rw, buffer[:size]) // Drain errors
				}

				return io.EOF
			}).
		Register(messaging.Error,
			func(buffer []byte, rw io.ReadWriter, size uint16) error {
				_, rErr := io.ReadFull(rw, buffer[:size])

				if rErr != nil {
					return rErr
				}

				return errors.New(
					"Server error message: " + string(buffer[:size]))
			}).
		Register(messaging.Closed,
			func(buffer []byte, rw io.ReadWriter, size uint16) error {
				if size > 0 {
					io.ReadFull(rw, buffer[:size]) // Drain errors
				}

				return ErrProcRemoteTargetClosed
			}).
		Register(messaging.InternalError,
			func(buffer []byte, rw io.ReadWriter, size uint16) error {
				if size > 0 {
					io.ReadFull(rw, buffer[:size]) // Drain errors
				}

				return ErrProcServerInternalError
			}).
		Register(messaging.Forbidden,
			func(buffer []byte, rw io.ReadWriter, size uint16) error {
				if size > 0 {
					io.ReadFull(rw, buffer[:size]) // Drain errors
				}

				return ErrProcServerRefused
			}).
		Register(messaging.Unconnectable,
			func(buffer []byte, rw io.ReadWriter, size uint16) error {
				if size > 0 {
					io.ReadFull(rw, buffer[:size]) // Drain errors
				}

				return ErrProcRemoteTargetUnconnectable
			}).
		Register(messaging.Timeout,
			func(buffer []byte, rw io.ReadWriter, size uint16) error {
				if size > 0 {
					io.ReadFull(rw, buffer[:size]) // Drain errors
				}

				return ErrProcTimeout
			}).
		Register(messaging.Unsupported,
			func(buffer []byte, rw io.ReadWriter, size uint16) error {
				if size > 0 {
					io.ReadFull(rw, buffer[:size]) // Drain errors
				}

				return ErrProcUnsupported
			}).
		Register(messaging.Invalid,
			func(buffer []byte, rw io.ReadWriter, size uint16) error {
				if size > 0 {
					io.ReadFull(rw, buffer[:size]) // Drain errors
				}

				return ErrProcInvalid
			}).
		Register(messaging.UnknownCommand,
			func(buffer []byte, rw io.ReadWriter, size uint16) error {
				if size > 0 {
					io.ReadFull(rw, buffer[:size]) // Drain errors
				}

				return ErrProcUnsupportedCommand
			})
}
