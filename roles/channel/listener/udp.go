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

	"github.com/nickrio/coward/common/codec"
	"github.com/nickrio/coward/common/locked"
	"github.com/nickrio/coward/roles/channel/common"
	"github.com/nickrio/coward/roles/channel/request"
	"github.com/nickrio/coward/roles/common/network/buffer"
	"github.com/nickrio/coward/roles/common/network/conn"
	dispatcher "github.com/nickrio/coward/roles/common/network/dispatcher/udp"
	"github.com/nickrio/coward/roles/common/network/transporter"
)

type udp struct {
	base

	listener  *net.UDPConn
	closeChan chan bool
}

// NewUDP creates a new UDP channel server
func NewUDP(config common.ServerConfig) (common.Server, error) {
	listenAddr, listenAddrErr := net.ResolveUDPAddr(
		"udp", net.JoinHostPort(config.Interface.String(),
			strconv.FormatUint(uint64(config.Port), 10)))

	if listenAddrErr != nil {
		return nil, listenAddrErr
	}

	listen, listenErr := net.ListenUDP("udp", listenAddr)

	if listenErr != nil {
		return nil, listenErr
	}

	return &udp{
		base: base{
			channel:     config.ID,
			defaultProc: config.DefaultProc,
			shutdown:    locked.NewBool(false),
			timeout:     config.Timeout,
			concurrence: config.Concurrence,
			transporter: config.Transporter,
			logger:      config.Logger,
		},
		listener:  listen,
		closeChan: make(chan bool),
	}, nil
}

// Serve starts server
func (u *udp) Serve(clientCloseWait *sync.WaitGroup) error {
	u.shutdown.Set(false)

	u.logger.Debugf("Serving")
	defer u.logger.Debugf("Closing")

	return dispatcher.New(func(client conn.UDPReadWriteCloser) error {
		selectedLogger := u.logger
		cancellerChan := make(transporter.Signal, 1)
		buf := buffer.Buffer{}

		clientCloseWait.Add(1)

		defer func() {
			client.Close()

			clientCloseWait.Done()
		}()

		_, requestErr := u.transporter.Request(
			request.NewUDPRequest(
				u.channel, client, u.defaultProc, u.closeChan),
			transporter.RequestOption{
				Canceller: cancellerChan,
				Buffer:    buf.Slice(),
				Delay: func(addr string, connectDelay float64, waiting uint64) {
					selectedLogger = u.logger.Context(addr)

					selectedLogger.Debugf("Transporter selected. Delay %f "+
						"seconds, %d requests are waiting",
						connectDelay, waiting)
				},
				Error: func(retry, reset bool, err error) (bool, bool, error) {
					if u.shutdown.Get() {
						return false, true, err
					}

					switch err.(type) {
					case *codec.Failure:
						selectedLogger.Warningf(
							"Decode error: %s. Retrying", err)

						return true, true, err
					}

					if retry {
						selectedLogger.Debugf("Error: %s. Retrying", err)
					}

					return retry, reset, err
				},
			})

		return requestErr
	}, u.concurrence, u.timeout).Dispatch(u.listener, func(err error) bool {
		return u.shutdown.Get()
	})
}

// Drop disconnects client from current listener
func (u *udp) Drop() {
	for {
		select {
		case u.closeChan <- true:
		default:
			return
		}
	}
}

// Close shut the server down
func (u *udp) Close(clientCloseWait *sync.WaitGroup) error {
	if u.shutdown.Get() {
		return common.ErrServerAlreadyClosed
	}

	defer u.logger.Debugf("Closed")

	u.shutdown.Set(true)

	closeErr := u.listener.Close()

	if closeErr != nil {
		return closeErr
	}

	clientCloseWait.Wait()

	close(u.closeChan)

	return nil
}
