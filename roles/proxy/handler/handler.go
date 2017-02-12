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

package handler

import (
	"io"
	"math/rand"
	"time"

	"github.com/nickrio/coward/common"
	"github.com/nickrio/coward/common/codec"
	"github.com/nickrio/coward/roles/common/network/buffer"
	"github.com/nickrio/coward/roles/common/network/messaging"
	"github.com/nickrio/coward/roles/common/network/transporter"
	pcommon "github.com/nickrio/coward/roles/proxy/common"
)

type handler struct {
	messaging.Messaging

	client         io.ReadWriter
	buffer         buffer.Slice
	proc           common.Proccessors
	channels       *pcommon.Channels
	connectTimeout time.Duration
	idleTimeout    time.Duration
	tempBuf        [8]byte
	resetTspConn   bool
	closeChan      chan bool
}

// NewHandler creates a new server handler
func NewHandler(
	config transporter.HandlerConfig,
	connectTimeout time.Duration,
	idleTimeout time.Duration,
	channels *pcommon.Channels,
	closeChan chan bool,
) transporter.Handler {
	h := &handler{
		client:         config.Server,
		buffer:         config.Buffer,
		proc:           nil,
		channels:       channels,
		connectTimeout: connectTimeout,
		idleTimeout:    idleTimeout,
		tempBuf:        [8]byte{},
		resetTspConn:   false,
		closeChan:      closeChan,
	}

	h.proc = common.NewProccessors().
		Register(messaging.NOP, h.nop).
		Register(messaging.RelayUDP, h.udp).
		Register(messaging.ConnectHost, h.connectHost).
		Register(messaging.ConnectIPv4, h.connectIPv4).
		Register(messaging.ConnectIPv6, h.connectIPv6).
		Register(messaging.ChannelTCP, h.channelTCP).
		Register(messaging.ChannelUDP, h.channelUDP)

	return h
}

func (h *handler) Handle() error {
	return h.Dispatch(h.client, h.buffer.Client.Buffer, h.proc)
}

func (h *handler) Error(err error) (bool, bool, error) {
	handleErr := err
	tspErr, isTSPErr := handleErr.(transporter.Error)

	if isTSPErr {
		handleErr = tspErr.Raw()
	}

	// Return the error to disconnect from client transporter
	// Return nil to keep serve the client transporter

	// If it's a decode error, send random bytes back so the
	// client can figure out that server is receiving invalid
	// data and retry connection after that

	switch e := handleErr.(type) {
	case codec.Error:
		randomLength := rand.Intn(len(h.buffer.Client.Buffer))

		_, rErr := rand.Read(h.buffer.Client.Buffer[:randomLength])

		if rErr != nil {
			h.Write(h.client, messaging.Error,
				[]byte(e.Error()), h.buffer.Client.ExtendedBuffer)
		} else {
			h.Write(h.client, messaging.Error,
				h.buffer.Client.Buffer[:randomLength],
				h.buffer.Client.ExtendedBuffer)
		}

		return false, false, err

	default:
		// For other error, check it and handle it accordingly
		switch e {
		case io.EOF:
			return false, h.resetTspConn, nil

		case common.ErrCommandUnsupported:
			h.Write(h.client, messaging.UnknownCommand, nil,
				h.buffer.Client.ExtendedBuffer)

		case ErrRequestingUndefindedChannel:
			h.Write(h.client, messaging.Unsupported, nil,
				h.buffer.Client.ExtendedBuffer)

		case ErrInvalidChannelID:
			fallthrough
		case ErrInvalidIPv6AddrPortLength:
			fallthrough
		case ErrInvalidIPv4AddrPortLength:
			fallthrough
		case ErrInvalidHostAddressPortLength:
			fallthrough
		case ErrDecodingPortBytes:
			h.Write(h.client, messaging.Invalid, nil,
				h.buffer.Client.ExtendedBuffer)

		case ErrHostNotFound:
			fallthrough
		case ErrDestinationUnconnectable:
			fallthrough
		case ErrInvalidAddress:
			h.Write(h.client, messaging.Unconnectable, nil,
				h.buffer.Client.ExtendedBuffer)

		case ErrLoopbackAddressIsForbidden:
			fallthrough
		case ErrZeroAddressIsForbidden:
			fallthrough
		case ErrZeroPortIsForbidden:
			h.Write(h.client, messaging.Forbidden, nil,
				h.buffer.Client.ExtendedBuffer)

		case ErrInvalidUDPEphemeralPortAddr:
			fallthrough
		case ErrFailedToOpenUDPEphemeralPort:
			h.Write(h.client, messaging.InternalError, nil,
				h.buffer.Client.ExtendedBuffer)

			return false, false, err

		case ErrFailedSendConnectConfirmSignal:
			return false, true, err

		default:
			// Default means don't send anything, or we may end up sending
			// multiple error feed back to user
		}
	}

	return false, h.resetTspConn, nil
}

func (h *handler) Close() error {
	return nil
}
