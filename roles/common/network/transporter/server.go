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
	"net"
	"time"

	"github.com/nickrio/coward/roles/common/network/transporter/common"
	"github.com/nickrio/coward/roles/common/network/transporter/connection"
)

// Server is what send and parse data
type Server interface {
	Serve(client net.Conn, builder HandlerBuilder, option ServeOption) error
}

// server implements Server
type server struct {
	base
}

// NewServer creates a new Transporter Server
func NewServer(
	connectionTimeout time.Duration,
	connectionPersistent bool,
	wrapper common.ConnWrapper,
	disrupter common.ConnDisrupter,
) Server {
	return &server{
		base: base{
			wrapper:              wrapper,
			disrupter:            disrupter,
			connectionPersistent: connectionPersistent,
			idleTimeout:          connectionTimeout,
		},
	}
}

// Serve serves the client connection
func (t *server) Serve(
	client net.Conn,
	builder HandlerBuilder,
	option ServeOption,
) error {
	var err error
	var handerErr error
	var serverErr error
	var resetTspConn bool

	wrappedConn, wrapErr := connection.WrapServerConn(
		client, t.wrapper, t.disrupter, t.idleTimeout)

	if wrapErr != nil {
		return wrapErr
	}

	handler := builder(HandlerConfig{
		Server: wrappedConn,
		Buffer: option.Buffer,
	})

	defer func() {
		handler.Close()

		// Close client connection AFTER handler closed
		// so we can still do comm in the handler
		client.Close()
	}()

	for {
		err = handler.Handle()

		if err == nil {
			continue
		}

		// Quit if it's a transporter error
		_, isTspErr := err.(*common.Errored)

		if isTspErr {
			return err
		}

		if option.Error != nil {
			// First, server will take look the error
			// If the error is serious enough, it will return
			// an error to close the connection
			serverErr = option.Error(err)

			if serverErr != nil {
				return serverErr
			}
		}

		// If every seems just fine by the server, the handler
		// will take look it again.
		// If the error is serious enough for handler, it will
		// return an error to close the connection
		// Server will not retry any if the request
		_, resetTspConn, handerErr = handler.Error(err)

		if handerErr != nil {
			return handerErr
		}

		if !t.connectionPersistent {
			return nil
		}

		if resetTspConn {
			return ErrHanderRequestedConnReset
		}
	}
}
