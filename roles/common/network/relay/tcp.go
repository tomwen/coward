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
	"net"

	"github.com/nickrio/coward/common"
	"github.com/nickrio/coward/roles/common/network/buffer"
	"github.com/nickrio/coward/roles/common/network/messaging"
)

// tcp is a TCP data exchanger
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
type tcp struct {
	messaging.Messaging

	terminal  net.Conn
	partner   io.ReadWriter
	buffer    buffer.Slice
	closeChan chan bool
}

// NewTCPRelay creates a new TCP relay
func NewTCPRelay(
	terminal net.Conn,
	partner io.ReadWriter,
	buffer buffer.Slice,
	closeChan chan bool,
) Relay {
	return &tcp{
		terminal:  terminal,
		partner:   partner,
		buffer:    buffer,
		closeChan: closeChan,
	}
}

// passTerminalStream handles terminal to partner data flow
func (t *tcp) passTerminalStream() error {
	for {
		terminalReadLen, terminalReadErr := t.terminal.Read(
			t.buffer.Client.Buffer)

		if terminalReadErr != nil {
			return terminalReadErr
		}

		_, partnerWriteErr := t.Write(t.partner, messaging.Streaming,
			t.buffer.Client.Buffer[:terminalReadLen],
			t.buffer.Server.ExtendedBuffer)

		if partnerWriteErr != nil {
			return partnerWriteErr
		}
	}
}

// handlePartnerStream handles partner to terminal data flow
func (t *tcp) handlePartnerStream() error {
	closing := false
	proc := common.NewProccessors().
		Register(messaging.Streaming, func(
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

			_, wErr := t.terminal.Write(b[:readLen])

			// Do not return wErr
			// handlePartnerStream meant to only return partner
			// related errors
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

			t.terminal.Close()

			if size <= 0 {
				return ErrTerminalConnectionClosed
			}

			io.ReadFull(rw, b[:size])

			return ErrTerminalConnectionClosed
		}).
		Register(messaging.EOF, func(
			b []byte,
			rw io.ReadWriter,
			size uint16,
		) error {
			closing = true

			if size <= 0 {
				return ErrPartnerCurrentSessionCompleted
			}

			io.ReadFull(rw, b[:size])

			return ErrPartnerCurrentSessionCompleted
		})

	for {
		partnerStreamDispatchErr := t.Dispatch(t.partner,
			t.buffer.Server.Buffer, proc)

		if partnerStreamDispatchErr == nil {
			continue
		}

		return partnerStreamDispatchErr
	}
}

// Relay exchanges data between partner and terminal
func (t *tcp) Relay() error {
	var resultErr error

	tmpBuf := [8]byte{}

	parter2terminalErrorChan := make(SignalChan)
	terminal2parterErrorChan := make(SignalChan)

	defer func() {
		close(parter2terminalErrorChan)
		close(terminal2parterErrorChan)
	}()

	go func() {
		parter2terminalErrorChan <- t.handlePartnerStream()
	}()

	go func() {
		terminal2parterErrorChan <- t.passTerminalStream()
	}()

	select {
	// Partner connection disconnected
	case p2dErr := <-parter2terminalErrorChan:
		resultErr = p2dErr

		t.terminal.Close()

		// Wait for d2p connection to complete
		<-terminal2parterErrorChan

		// Don't send Unleashed signal in a loop
		switch p2dErr {
		case ErrPartnerCurrentSessionCompleted:
			// Do nothing

		case ErrTerminalConnectionClosed:
			t.Write(t.partner, messaging.EOF, nil,
				t.buffer.Server.ExtendedBuffer)
		}

	// Terminal connection disconnected
	case d2pErr := <-terminal2parterErrorChan:
		resultErr = d2pErr

		t.Write(t.partner, messaging.Closed, nil,
			t.buffer.Server.ExtendedBuffer)

		// Wait for p2d connection to complete
		<-parter2terminalErrorChan

	case <-t.closeChan:
		t.Write(t.partner, messaging.Closed, nil,
			tmpBuf[:])

		<-parter2terminalErrorChan

		t.terminal.Close()

		<-terminal2parterErrorChan
	}

	return resultErr
}
