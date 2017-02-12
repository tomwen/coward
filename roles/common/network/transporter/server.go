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
	"sync"

	"github.com/nickrio/coward/common/locked"
)

// Transporter server errors
var (
	ErrServerNotActive = errors.New(
		"Transporter server is not active")
)

// Server is what receive and handle data request
type Server interface {
	Serve(option ServeOptionBuilder, accepter chan ServerConnAccepterMeta) error
	Close() error
}

// server implements Server
type server struct {
	listener     ServerConnListener
	accepter     ServerConnAccepter
	reuseConn    bool
	shutdown     locked.Boolean
	shutdownWait sync.WaitGroup
}

// NewServer creates a new Transporter Server
func NewServer(listener ServerConnListener, reuseConn bool) Server {
	return &server{
		listener:     listener,
		reuseConn:    reuseConn,
		shutdown:     locked.NewBool(false),
		shutdownWait: sync.WaitGroup{},
	}
}

// handle handles incoming requests
func (s *server) handle(
	clientConn ServerClientConn,
	option ServeOption,
) error {
	var err error
	var isTSPErr bool
	var needBreak bool
	var needDisconnect bool

	option.Connected(clientConn)

	handler := option.Handler(HandlerConfig{
		Server: &wrapped{ReadWriteCloser: clientConn},
		Buffer: option.Buffer,
	})

	defer func() {
		handler.Close()

		// Close client connection AFTER handler closed
		// so we can still do comm in the handler
		clientConn.Close()

		option.Disconnected(clientConn, err)
	}()

	for {
		err = handler.Handle()

		// If current error is a Transporter error, meaning
		// current Transporter connection is not working, thus
		// needs to be disconnected
		_, isTSPErr = err.(Error)

		if isTSPErr {
			needBreak = true
		}

		// Let handler filter the error, so we can decide whether
		// or not we will disconnect current connection
		_, needDisconnect, err = handler.Error(err)

		if err != nil && needDisconnect {
			needBreak = true
		}

		// option.Error can cancel break
		err = option.Error(err)

		if err == nil && !isTSPErr {
			needBreak = false
		}

		// Will we break or continue?
		if !needBreak && s.reuseConn {
			continue
		}

		break
	}

	return err
}

func (s *server) serve(opt ServeOptionBuilder) error {
	var accept ServerClientConn
	var acceptErr error

	for {
		accept, acceptErr = s.accepter.Accept()

		if acceptErr != nil {
			if s.shutdown.Get() {
				break
			}

			continue
		}

		s.shutdownWait.Add(1)

		go func(conn ServerClientConn) {
			defer s.shutdownWait.Done()

			s.handle(conn, opt(conn))
		}(accept)
	}

	return acceptErr
}

// Listen to the listner
func (s *server) Serve(
	optionBuilder ServeOptionBuilder,
	accepter chan ServerConnAccepterMeta,
) error {
	var accept ServerConnAccepter
	var acceptErr error

	s.shutdownWait.Wait()

	accept, acceptErr = s.listener.Listen()

	if acceptErr != nil {
		// tell we failed
		accepter <- nil

		return acceptErr
	}

	// tell we ready
	accepter <- accept

	s.accepter = accept

	defer func() {
		s.accepter = nil
	}()

	return s.serve(optionBuilder)
}

// Close closes current server is there is any
func (s *server) Close() error {
	if s.accepter == nil {
		return ErrServerNotActive
	}

	s.shutdown.Set(true)

	closeErr := s.accepter.Close()

	s.shutdownWait.Wait()

	s.accepter = nil

	return closeErr
}
