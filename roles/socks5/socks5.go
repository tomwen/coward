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

package socks5

import (
	"net"
	"strconv"
	"sync"
	"time"

	ccommon "github.com/nickrio/coward/common"
	"github.com/nickrio/coward/common/codec"
	"github.com/nickrio/coward/common/logger"
	"github.com/nickrio/coward/common/role"
	"github.com/nickrio/coward/roles/common/network"
	"github.com/nickrio/coward/roles/common/network/buffer"
	"github.com/nickrio/coward/roles/common/network/conn"
	"github.com/nickrio/coward/roles/common/network/transporter"
	"github.com/nickrio/coward/roles/common/network/transporter/balancer"
	"github.com/nickrio/coward/roles/socks5/common"
	"github.com/nickrio/coward/roles/socks5/defaults"
)

// socks5 is a partially compatible implementation of RFC1928
type socks5 struct {
	connector    balancer.Balancer
	config       Config
	serverWaiter sync.WaitGroup
	listener     net.Listener
	atypeStream  common.ATYP
	atypeBlock   common.Address
	auther       common.Auther
	proc         ccommon.Proccessors
	shuttingDown bool
	closeNotify  chan<- bool
}

// New creates a new Socks5 proxy
func New(conn balancer.Balancer, cfg Config) role.Role {
	s5 := &socks5{
		connector:    conn,
		config:       cfg,
		serverWaiter: sync.WaitGroup{},
		atypeStream:  defaults.GetATYPStream(),
		atypeBlock:   defaults.GetATYPBlock(),
		auther:       defaults.GetAuther(cfg.Auth),
		proc:         network.GetDefaultProc(),
	}

	return s5
}

func (s *socks5) Spawn(closeNotify chan<- bool) error {
	listen, listenErr := net.Listen("tcp", net.JoinHostPort(
		s.config.Interface.String(),
		strconv.FormatUint(uint64(s.config.Port), 10)))

	if listenErr != nil {
		s.config.Logger.Errorf(
			"Can't start server due to error: %s", listenErr)

		return listenErr
	}

	s.listener = listen

	s.serverWaiter.Add(1)

	go func() {
		defer s.serverWaiter.Done()

		s.serve()
	}()

	s.config.Logger.Infof(
		"Server is up, listening %s", listen.Addr().String())

	s.closeNotify = closeNotify
	s.shuttingDown = false

	return nil
}

func (s *socks5) Unspawn() error {
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

	s.config.Logger.Infof("Closing connections")

	s.serverWaiter.Wait()

	s.config.Logger.Infof("Server is down")

	return nil
}

func (s *socks5) serve() {
	keepKicking := true
	shutdownWait := sync.WaitGroup{}
	clientWait := sync.WaitGroup{}
	connections := network.NewConnections(256)

	shutdownWait.Add(1)

	defer func() {
		go func() {
			defer shutdownWait.Done()

			for keepKicking {
				connections.Iterate(func(name string, conn net.Conn) {
					conn.Close()
				})

				s.connector.Kickoff()
			}
		}()

		clientWait.Wait()

		keepKicking = false

		shutdownWait.Wait()
	}()

	for {
		client, acceptErr := s.listener.Accept()

		if acceptErr != nil {
			if s.shuttingDown {
				break
			}

			s.config.Logger.Errorf(
				"Can't accpet connection due to error: %s", acceptErr)

			time.Sleep(1 * time.Second)

			continue
		}

		name := client.RemoteAddr().String()

		connections.Put(name, client)

		clientWait.Add(1)

		go func(name string, c net.Conn) {
			defer func() {
				connections.Del(name)

				// Ignore the error if there is any.
				// We'll try to close it anyway
				c.Close()

				clientWait.Done()
			}()

			var handleErr error

			cLog := s.config.Logger.Context(c.RemoteAddr().String())

			defer func() {
				if handleErr == nil {
					cLog.Debugf("Disconnected")

					return
				}

				switch handleErr {
				case common.ErrAuthFailed:
					cLog.Warningf("Disconnected: %s", handleErr)

				default:
					cLog.Debugf("Disconnected: %s", handleErr)
				}
			}()

			cLog.Debugf("Connected")

			handleErr = s.handle(c, cLog)
		}(name, client)
	}
}

// handle handles Socks 5 requests
func (s *socks5) handle(client net.Conn, log logger.Logger) error {
	var err error

	cancellerChan := make(transporter.Signal)
	buf := buffer.Buffer{}

	wrappedClient := conn.WrapClientConn(client, conn.ClientConfig{
		Timeout: s.config.Timeout,
		OnClose: func() {
			select {
			case cancellerChan <- nil:
			default:
			}
		},
	})

	defer wrappedClient.Close()

	n := negotiator{
		auther:      &s.auther,
		atypeStream: &s.atypeStream,
		atypeBlock:  &s.atypeBlock,
		proc:        s.proc,
		cmd:         common.Command{},
		current:     handshake,
		buffer:      buf.Client.Buffer[:],
		steps:       [3]func(net.Conn) error{},
		request: func(
			addr string,
			builder balancer.DelayFeedingbackRequestBuilder,
		) error {
			return s.connector.Request(addr, builder, transporter.RequestOption{
				Canceller: cancellerChan,
				Buffer:    buf.Slice(),
				Delay:     func(connectDelay float64, wait uint64) {},
				Error: func(retry, reset bool, err error) (bool, bool, error) {
					if s.shuttingDown {
						return false, true, err
					}

					switch opte := err.(type) {
					case transporter.Error:
						switch e := opte.Raw().(type) {
						case codec.Error:
							log.Warningf("Decode error: %s. Retrying", e)

							return true, true, err
						}

					case conn.ErrorConnError:
						retry = false
					}

					if retry {
						log.Debugf("Error: %s. Retrying", err)
					}

					return retry, reset, err
				},
			})
		},
	}

	n.Inital()

	for {
		err = n.Loop(wrappedClient)

		if err != nil {
			break
		}
	}

	return err
}
