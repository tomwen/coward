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

import (
	"errors"

	"github.com/nickrio/coward/roles/common/network/buffer"
)

// Errors for both transporter server and client
var (
	ErrHanderRequestedConnReset = errors.New(
		"Disconnected in response to handlers request")
)

// Signal a type of channel that will be used to send signals
type Signal chan error

// RequestOption is the configuration for client request
type RequestOption struct {
	Buffer    buffer.Slice
	Canceller Signal
	Delay     func(addr string, avgConnectDelay float64, waiting uint64)
	Error     func(
		wantToRetry bool,
		wantToResetTransportConn bool,
		executeErr error,
	) (retry bool, reset bool, err error)
}

// ServeOption is the configuration for server handler
type ServeOption struct {
	Buffer buffer.Slice
	Error  func(error) error
}
