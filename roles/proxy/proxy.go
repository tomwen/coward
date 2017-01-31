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
	"net"
	"strconv"
	"sync"

	"github.com/nickrio/coward/common/codec"
	"github.com/nickrio/coward/common/logger"
	"github.com/nickrio/coward/common/role"
	"github.com/nickrio/coward/roles/common/network"
	"github.com/nickrio/coward/roles/common/network/buffer"
	"github.com/nickrio/coward/roles/common/network/transporter"
	"github.com/nickrio/coward/roles/proxy/handler"
)

type proxy struct {
	transporter  transporter.Server
	config       Config
	clientWaiter sync.WaitGroup
	serverWaiter sync.WaitGroup
	listener     net.Listener
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
		listener:     nil,
		shuttingDown: false,
		closeNotify:  nil,
	}

	return p
}

func (s *proxy) Spawn(closeNotify chan<- bool) error {
	listen, listenErr := net.Listen("tcp", net.JoinHostPort(
		s.config.Interface.String(),
		strconv.FormatUint(uint64(s.config.Port), 10)))

	if listenErr != nil {
		s.config.Logger.Errorf("Can't listen due to error %s", listenErr)

		return listenErr
	}

	s.listener = listen

	s.serverWaiter.Add(1)

	go func() {
		defer s.serverWaiter.Done()

		s.serve()
	}()

	s.config.Logger.Infof(
		"Server is up, listening %s", s.listener.Addr())

	s.closeNotify = closeNotify
	s.shuttingDown = false

	return nil
}

func (s *proxy) Unspawn() error {
	s.shuttingDown = true // Set flag before actually close

	closeErr := s.listener.Close()

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

func (s *proxy) serve() {
	shutdownWait := sync.WaitGroup{}
	forceCloseChan := make(chan bool)
	connections := network.NewConnections()

	defer func() {
		breakLoop := false

		shutdownWait.Add(1)

		go func() {
			defer func() {
				close(forceCloseChan)

				connections.Close()

				shutdownWait.Done()
			}()

			for !breakLoop {
				select {
				case forceCloseChan <- true:
				default:
					connections.Iterate(func(name string, conn net.Conn) {
						conn.Close()
					})
				}
			}
		}()

		s.clientWaiter.Wait()

		breakLoop = true

		shutdownWait.Wait()
	}()

	shutdownWait.Add(1)

	go func() {
		defer shutdownWait.Done()

		connections.Serve()
	}()

	for {
		client, acceptErr := s.listener.Accept()

		if acceptErr != nil {
			if s.shuttingDown {
				break
			}

			s.config.Logger.Errorf(
				"Can't accpet connection due to error: %s", acceptErr)

			continue
		}

		name := client.RemoteAddr().String()

		connections.Put(name, client)

		s.clientWaiter.Add(1)

		go func(name string, c net.Conn) {
			defer func() {
				connections.Del(name)

				// Ignore the error if there is any.
				// We'll try to close it anyway
				c.Close()

				s.clientWaiter.Done()
			}()

			cLog := s.config.Logger.Context(
				c.RemoteAddr().String())

			var handleErr error

			defer func() {
				if handleErr != nil {
					cLog.Debugf("Disconnected: %s", handleErr)
				} else {
					cLog.Debugf("Disconnected")
				}
			}()

			cLog.Debugf("Connected")

			handleErr = s.handle(c, cLog, forceCloseChan)
		}(name, client)
	}
}

func (s *proxy) handle(
	client net.Conn, log logger.Logger, closeChan chan bool) error {
	buf := buffer.Buffer{}

	return s.transporter.Serve(
		client, func(hc transporter.HandlerConfig) transporter.Handler {
			return handler.NewHandler(hc, s.config.ConnectTimeout,
				s.config.IdleTimeout, &s.config.Channels, closeChan)
		}, transporter.ServeOption{
			Error: func(er error) error {
				switch er.(type) {
				case *codec.Failure:
					log.Warningf("Decode error: %s", er)
				default:
					log.Debugf("General error: %s", er)
				}

				return nil
			},
			Buffer: buf.Slice(),
		},
	)
}
