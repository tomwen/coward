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

package listener

import (
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/nickrio/coward/common/codec"
	"github.com/nickrio/coward/common/locked"
	"github.com/nickrio/coward/common/logger"
	"github.com/nickrio/coward/roles/channel/common"
	"github.com/nickrio/coward/roles/channel/request"
	"github.com/nickrio/coward/roles/common/network"
	"github.com/nickrio/coward/roles/common/network/buffer"
	"github.com/nickrio/coward/roles/common/network/conn"
	"github.com/nickrio/coward/roles/common/network/transporter"
)

type tcp struct {
	base

	listener    net.Listener
	connections network.Connections
}

// NewTCP creates a new TCP channel listener
func NewTCP(config common.ServerConfig) (common.Server, error) {
	listenAddr, listenAddrErr := net.ResolveTCPAddr(
		"tcp", net.JoinHostPort(config.Interface.String(),
			strconv.FormatUint(uint64(config.Port), 10)))

	if listenAddrErr != nil {
		return nil, listenAddrErr
	}

	listen, listenErr := net.ListenTCP("tcp", listenAddr)

	if listenErr != nil {
		return nil, listenErr
	}

	return &tcp{
		base: base{
			channel:     config.ID,
			defaultProc: config.DefaultProc,
			shutdown:    locked.NewBool(false),
			timeout:     config.Timeout,
			concurrence: config.Concurrence,
			transporter: config.Transporter,
			logger:      config.Logger,
		},
		listener:    listen,
		connections: network.NewConnections(256),
	}, nil
}

// handle handles incomming requests
func (t *tcp) handle(client net.Conn, log logger.Logger) error {
	cancellerChan := make(transporter.Signal, 1)
	buf := buffer.Buffer{}

	wrappedClient := conn.WrapClientConn(client, conn.ClientConfig{
		Timeout: t.timeout,
		OnClose: func() {
			select {
			case cancellerChan <- nil:
			default:
			}
		},
	})

	defer wrappedClient.Close()

	_, requestErr := t.transporter.Request(
		request.NewTCPRequest(t.channel, client, t.defaultProc),
		transporter.RequestOption{
			Canceller: cancellerChan,
			Buffer:    buf.Slice(),
			Delay:     func(connectDelay float64, waiting uint64) {},
			Error: func(retry, reset bool, err error) (bool, bool, error) {
				if t.shutdown.Get() {
					return false, true, err
				}

				switch ee := err.(type) {
				case transporter.Error:
					switch eee := ee.Raw().(type) {
					case codec.Error:
						log.Warningf(
							"Decode error: %s. Retrying", eee)

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

	return requestErr
}

// Serve starts the server
func (t *tcp) Serve(clientCloseWait *sync.WaitGroup) error {
	concurrentLimitChan := make(chan bool, t.concurrence)

	t.logger.Debugf("Serving")
	defer t.logger.Debugf("Closing")

	for ccLoop := uint16(0); ccLoop < t.concurrence; ccLoop++ {
		concurrentLimitChan <- true
	}

	for {
		<-concurrentLimitChan

		listener, listenerErr := t.listener.Accept()

		if listenerErr != nil {
			if t.shutdown.Get() {
				break
			}

			t.logger.Warningf("Can't accept connection due to error: %s",
				listenerErr)

			time.Sleep(1 * time.Second)

			continue
		}

		name := listener.RemoteAddr().String()

		t.connections.Put(name, listener)

		clientCloseWait.Add(1)

		go func(name string, l net.Conn) {
			var err error

			log := t.logger.Context(listener.RemoteAddr().String())

			defer func() {
				if err != nil {
					log.Debugf("Disconnected: %s", err)
				} else {
					log.Debugf("Disconnected")
				}

				t.connections.Del(name)

				l.Close()

				concurrentLimitChan <- true

				clientCloseWait.Done()
			}()

			err = t.handle(l, log)
		}(name, listener)
	}

	return nil
}

// Drop disconnects client from current listener
func (t *tcp) Drop() {
	t.connections.Iterate(func(name string, conn net.Conn) {
		conn.Close()
	})
}

// Close shutdown current server
func (t *tcp) Close(clientCloseWait *sync.WaitGroup) error {
	if t.shutdown.Get() {
		return common.ErrServerAlreadyClosed
	}

	defer t.logger.Debugf("Closed")

	t.shutdown.Set(true)

	closeErr := t.listener.Close()

	if closeErr != nil {
		return closeErr
	}

	clientCloseWait.Wait()

	return nil
}
