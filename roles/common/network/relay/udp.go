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

package relay

import (
	"io"

	"github.com/nickrio/coward/common"
	"github.com/nickrio/coward/roles/common/network/buffer"
	"github.com/nickrio/coward/roles/common/network/conn"
	"github.com/nickrio/coward/roles/common/network/messaging"
)

// udp is a UDP data exchanger
//
// Notice:
//
// - Here is how we use buffers:
//
//   Partner: Read:
//                t.buffer.Server.Buffer
//            Write:
//                t.buffer.Server.ExtendedBuffer
//
//   Terminal: Read:
//                t.buffer.Client.Buffer
//            Write:
//                t.buffer.Client.ExtendedBuffer
//
type udp struct {
	messaging.Messaging

	handler   UDPHandler
	terminal  conn.UDPReadWriteCloser
	partner   io.ReadWriter
	buffer    buffer.Slice
	closeChan chan bool
}

// NewUDPRelay creates a new UDP relay
func NewUDPRelay(
	handler UDPHandler,
	terminal conn.UDPReadWriteCloser,
	partner io.ReadWriter,
	buffer buffer.Slice,
	closeChan chan bool,
) Relay {
	return &udp{
		handler:   handler,
		terminal:  terminal,
		partner:   partner,
		buffer:    buffer,
		closeChan: closeChan,
	}
}

func (u *udp) passTerminalPacket() error {
	clientBufferLen := len(u.buffer.Client.Buffer)

	for {
		exReadLen, exReadErr := u.handler.Receive(
			u.terminal,
			u.buffer.Client.Buffer,
			u.buffer.Server.ExtendedBuffer[:clientBufferLen])

		if exReadErr != nil {
			return exReadErr
		}

		_, partnerWriteErr := u.Write(u.partner, messaging.Datagram,
			u.buffer.Client.Buffer[:exReadLen],
			u.buffer.Server.ExtendedBuffer)

		if partnerWriteErr != nil {
			return partnerWriteErr
		}
	}
}

func (u *udp) handlePartnerStream() error {
	closing := false
	proc := common.NewProccessors().
		Register(messaging.Datagram, func(
			b []byte,
			rw io.ReadWriter,
			size uint16,
		) error {
			readLen, readErr := io.ReadFull(rw, b[:size])

			if readErr != nil {
				return readErr
			}

			if closing {
				return nil
			}

			_, wErr := u.handler.Send(u.terminal, b[:readLen],
				u.buffer.Client.ExtendedBuffer)

			if wErr != nil {
				closing = true
			}

			return nil
		}).
		Register(messaging.Closed, func(
			b []byte,
			rw io.ReadWriter,
			size uint16,
		) error {
			closing = true

			u.terminal.Close()

			if size <= 0 {
				return ErrTerminalConnectionClosed
			}

			io.ReadFull(rw, b[:size])

			return ErrTerminalConnectionClosed
		}).
		Register(messaging.Unleash, func(
			b []byte,
			rw io.ReadWriter,
			size uint16,
		) error {
			if size <= 0 {
				return ErrPartnerConnectionUnleashed
			}

			io.ReadFull(rw, b[:size])

			return ErrPartnerConnectionUnleashed
		}).
		Register(messaging.EOF, func(
			b []byte,
			rw io.ReadWriter,
			size uint16,
		) error {
			if size <= 0 {
				return ErrPartnerCurrentSessionCompleted
			}

			io.ReadFull(rw, b[:size])

			return ErrPartnerCurrentSessionCompleted
		})

	for {
		partnerStreamDispatchErr := u.Dispatch(u.partner,
			u.buffer.Server.Buffer, proc)

		if partnerStreamDispatchErr == nil {
			continue
		}

		return partnerStreamDispatchErr
	}
}

func (u *udp) Relay() error {
	var resultErr error
	tmpBuf := [8]byte{}

	readyErr := u.handler.Ready()

	if readyErr != nil {
		return readyErr
	}

	parter2terminalErrorChan := make(SignalChan)
	terminal2parterErrorChan := make(SignalChan)

	defer func() {
		close(parter2terminalErrorChan)
		close(terminal2parterErrorChan)
	}()

	go func() {
		parter2terminalErrorChan <- u.handlePartnerStream()
	}()

	go func() {
		terminal2parterErrorChan <- u.passTerminalPacket()
	}()

	select {
	// Partner connection disconnected
	case resultErr = <-parter2terminalErrorChan:
		u.terminal.Close()

		<-terminal2parterErrorChan

		switch resultErr {
		case ErrPartnerCurrentSessionCompleted:
			// Do nothing
		case ErrPartnerConnectionUnleashed:
			u.Write(u.partner, messaging.EOF, nil,
				u.buffer.Server.ExtendedBuffer)
		default:
			u.Write(u.partner, messaging.Unleash, nil,
				u.buffer.Server.ExtendedBuffer)
		}

	// Terminal connection disconnected
	case resultErr = <-terminal2parterErrorChan:
		u.Write(u.partner, messaging.Closed, nil,
			u.buffer.Server.ExtendedBuffer)

		// Wait for p2d connection to complete
		<-parter2terminalErrorChan

	// Quitter chan
	case resultErr = <-u.handler.Quitter():
		u.Write(u.partner, messaging.Unleash, nil, tmpBuf[:])

		<-parter2terminalErrorChan

		u.terminal.Close()

		<-terminal2parterErrorChan

	// Close chan
	case <-u.closeChan:
		u.Write(u.partner, messaging.Unleash, nil, tmpBuf[:])

		<-parter2terminalErrorChan

		u.terminal.Close()

		<-terminal2parterErrorChan
	}

	return resultErr
}
