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

package proxy

import (
	"sync"

	"github.com/nickrio/coward/common/codec"
	"github.com/nickrio/coward/common/logger"
	"github.com/nickrio/coward/common/role"
	"github.com/nickrio/coward/roles/common/network/buffer"
	"github.com/nickrio/coward/roles/common/network/transporter"
	"github.com/nickrio/coward/roles/proxy/handler"
)

type proxy struct {
	transporter  transporter.Server
	config       Config
	clientWaiter sync.WaitGroup
	serverWaiter sync.WaitGroup
	shuttingDown bool
	closeNotify  chan<- bool
}

// New creates a new COWARD Encoded Proxy server
func New(
	transp transporter.Server,
	cfg Config,
) role.Role {
	p := &proxy{
		transporter:  transp,
		config:       cfg,
		clientWaiter: sync.WaitGroup{},
		serverWaiter: sync.WaitGroup{},
		shuttingDown: false,
		closeNotify:  nil,
	}

	return p
}

func (s *proxy) Spawn(closeNotify chan<- bool) error {
	var listenErr error

	acceptMeta := make(chan transporter.ServerConnAccepterMeta)

	s.serverWaiter.Add(1)

	go func() {
		defer s.serverWaiter.Done()

		listenErr = s.serve(s.config.Logger, acceptMeta)
	}()

	accepterInfo := <-acceptMeta

	if accepterInfo == nil {
		// Wait the Transporter Server down before continuing,
		// because s.serve may returned AFTER acceptMeta been filled
		// If we don't wait, listenErr may contains nothing
		s.serverWaiter.Wait()

		if listenErr != nil {
			s.config.Logger.Errorf("Can't listen due to error %s", listenErr)

			return listenErr
		}
	}

	s.config.Logger.Infof(
		"Server is up, listening %s", accepterInfo.Name())

	s.closeNotify = closeNotify
	s.shuttingDown = false

	return nil
}

func (s *proxy) Unspawn() error {
	s.shuttingDown = true // Set flag before actually close

	closeErr := s.transporter.Close()

	if closeErr != nil {
		s.config.Logger.Errorf(
			"Can't close server due to error: %s", closeErr)

		return closeErr
	}

	defer func() {
		s.closeNotify <- true
	}()

	s.config.Logger.Infof("Waiting port to be closed")

	s.serverWaiter.Wait()

	s.config.Logger.Infof("Server is down")

	return nil
}

func (s *proxy) serve(
	log logger.Logger,
	acceptInfo chan transporter.ServerConnAccepterMeta,
) error {
	return s.transporter.Serve(
		func(client transporter.ServerClientInfo) transporter.ServeOption {
			buf := buffer.Buffer{}
			clientLog := log.Context(client.Name())

			return transporter.ServeOption{
				Buffer: buf.Slice(),
				Handler: func(
					hc transporter.HandlerConfig) transporter.Handler {
					return handler.NewHandler(hc, s.config.ConnectTimeout,
						s.config.IdleTimeout, &s.config.Channels, nil)
				},
				Connected: func(clientInfo transporter.ServerClientInfo) {
					clientLog.Debugf("Connected")
				},
				Disconnected: func(
					clientInfo transporter.ServerClientInfo, err error) {
					if err != nil {
						clientLog.Debugf("Disconnected: ", err)

						return
					}

					clientLog.Debugf("Disconnected")
				},
				Error: func(er error) error {
					if er == nil {
						return nil
					}

					switch e := er.(type) {
					case transporter.Error:
						// If it's a Transporter error
						switch tspErr := e.Raw().(type) {
						case codec.Error:
							clientLog.Warningf("Decode error: %s", tspErr)
						}

					default:
						clientLog.Debugf("General error: %s", er)
					}

					return nil
				},
			}
		}, acceptInfo)
}
