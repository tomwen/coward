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
	"errors"
	"io"
	"net"
	"time"

	ccommon "github.com/nickrio/coward/common"
	"github.com/nickrio/coward/roles/common/network/transporter"
	"github.com/nickrio/coward/roles/common/network/transporter/balancer"
	"github.com/nickrio/coward/roles/socks5/common"
	"github.com/nickrio/coward/roles/socks5/request"
)

type negotiation byte

// negotiator errors
var (
	ErrRequestCompleted = errors.New(
		"Request completed")
)

const (
	handshake negotiation = 0
	connect   negotiation = 1
	finish    negotiation = 2
)

type negotiator struct {
	auther      *common.Auther
	atypeStream *common.ATYP
	atypeBlock  *common.Address
	proc        ccommon.Proccessors
	cmd         common.Command
	current     negotiation
	buffer      []byte
	steps       [3]func(net.Conn) error
	request     func(string, balancer.DelayFeedingbackRequestBuilder) error
}

func (n *negotiator) Inital() {
	n.current = handshake

	n.cmd.Register(common.Connect,
		func(
			aType common.ATYPE,
			addr []byte,
			port []byte,
			target []byte,
			rw net.Conn,
		) error {
			return n.request(
				"TCP"+string(target),
				func(
					cfg transporter.HandlerConfig,
					delayBack func(time.Duration),
				) transporter.Handler {
					return request.NewConnectRequest(cfg, n.proc, rw, aType,
						addr, port, delayBack)
				})
		})

	n.cmd.Register(common.UDP,
		func(
			aType common.ATYPE,
			addr []byte,
			port []byte,
			target []byte,
			rw net.Conn,
		) error {
			return n.request(
				"UDP"+string(target),
				func(
					cfg transporter.HandlerConfig,
					delayBack func(time.Duration),
				) transporter.Handler {
					return request.NewUDPRequest(cfg, n.proc, n.atypeBlock, rw,
						aType, addr, port, delayBack)
				})
		})

	// Steps
	n.steps = [3]func(net.Conn) error{
		// Handshake
		func(rw net.Conn) error {
			isAuthed := false

			// Handshake message format:
			//
			// +----+----------+----------+
			// |VER | NMETHODS | METHODS  |
			// +----+----------+----------+
			// | 1  |    1     | 1 to 255 |
			// +----+----------+----------+
			//
			//
			// Failed message format:
			//
			// +----+--------+
			// |VER | METHOD |
			// +----+--------+
			// | 1  |   1    |
			// +----+--------+

			// Read VER and NMETHODS
			rLen, rErr := io.ReadFull(rw, n.buffer[:2])

			if rErr != nil || rLen < 2 {
				return common.ErrFailedToReadHandshakeHead
			}

			if n.buffer[0] != common.Version {
				return common.ErrUnsupportedSocksVersion
			}

			if n.buffer[1] == 0 {
				return common.ErrNoAuthMethodProvided
			}

			authMethods := n.buffer[1]

			// Read METHODS
			rLen, rErr = io.ReadFull(rw, n.buffer[:authMethods])

			if rErr != nil {
				return common.ErrFailedToReadAllAuthMethods
			}

			// Do auth
			var lastAuthErr error

			for _, method := range n.buffer[:rLen] {
				if !n.auther.Has(common.AuthMethod(method)) {
					continue
				}

				authErr := n.auther.Auth(
					common.AuthMethod(method), rw, n.buffer[:])

				if authErr != nil {
					lastAuthErr = authErr

					continue
				}

				isAuthed = true

				break
			}

			// No auth method? Tell client we can't handle this auth
			if !isAuthed {
				if lastAuthErr != nil {
					return common.ErrAuthFailed
				}

				n.buffer[0] = common.Version
				n.buffer[1] = byte(common.Unauthable)

				_, wErr := rw.Write(n.buffer[:2])

				if wErr != nil {
					return wErr
				}

				return common.ErrUnsupportedAuthMethod
			}

			n.current = connect

			return nil
		},

		// Request handler
		func(rw net.Conn) error {
			// Request format:
			//
			// +----+-----+-------+------+----------+----------+
			// |VER | CMD |  RSV  | ATYP | DST.ADDR | DST.PORT |
			// +----+-----+-------+------+----------+----------+
			// | 1  |  1  | X'00' |  1   | Variable |    2     |
			// +----+-----+-------+------+----------+----------+

			// Read VER, CMD, RSV, ATYP
			_, rErr := io.ReadFull(rw, n.buffer[:4])

			if rErr != nil {
				return common.ErrFailedToReadRequestHead
			}

			if n.buffer[0] != common.Version {
				return common.ErrUnsupportedSocksVersion
			}

			cmd := n.buffer[1]

			// Get the address and host. Notice we still using same
			// buffer space here, so we have to read by offset
			addrType, addr, addErr := n.atypeStream.Run(
				common.ATYPE(n.buffer[3]),
				rw,
				n.buffer[:])

			if addErr != nil {
				// We don't return any error if the address type is
				// invalid, just close the connection let client
				// figure it out by itself.

				// I know, we should return the right errors to tell
				// the client what's wrong. But that require me to
				// change the stream reader, which is something that
				// takes a lots of time.
				// So while benefit of doing so is so tiny, I think
				// it's easy for me to make decision to not do that

				return addErr
			}

			// Max addrLen is 255, and buffer size is about 4096
			addrLen := len(addr)

			// So we can do this without check remaining size
			_, rErr = io.ReadFull(rw, n.buffer[addrLen:addrLen+2])

			if rErr != nil {
				return rErr
			}

			// We do this far, now we will go back whatever the command
			// is succeed or failed. This will reset the state of current
			// connection.
			n.current = finish

			cmdErr := n.cmd.Run(
				common.CTYPE(cmd),
				addrType,
				addr,
				n.buffer[addrLen:addrLen+2],
				n.buffer[:addrLen+2],
				rw,
			)

			if cmdErr != nil {
				return cmdErr
			}

			return nil
		},

		// Finish request and disconnect
		func(rw net.Conn) error {
			return ErrRequestCompleted
		},
	}
}

func (n *negotiator) Loop(rw net.Conn) error {
	currentErr := n.steps[n.current](rw)

	if currentErr != nil {
		return currentErr
	}

	return nil
}
