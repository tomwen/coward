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

import "github.com/nickrio/coward/roles/common/network/transporter/common"

// Transporter connection errors
var (
	ErrTimeout = common.Error(
		"Transporter connection timeout")

	ErrReadTimeout = common.Error(
		"Transporter connection read timeout")

	ErrWriteTimeout = common.Error(
		"Transporter connection write timeout")

	ErrUnconnectable = common.Error(
		"Transporter host is unconnectable")

	ErrRefused = common.Error(
		"Transporter connection refused")

	ErrResetted = common.Error(
		"Transporter connection resetted")

	ErrAborted = common.Error(
		"Transporter connection aborted")

	ErrEOF = common.Error(
		"Transporter connection terminated")
)

var (
	// ErrBroken is throwed when a transporter connection is
	// failed during the first attempt of use.
	// Which means that the transporter connection is never
	// able to transmit or recevie any data for the current
	// request handler
	ErrBroken = common.Error(
		"Transporter connection is broken and can't be used")

	// ErrReclosing is throwed when trying to close a connection
	// which already been closed
	ErrReclosing = common.Error(
		"Closing an already closed transporter connection")
)
