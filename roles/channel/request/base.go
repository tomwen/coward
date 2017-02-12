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

package request

import (
	"io"

	"github.com/nickrio/coward/common"
	"github.com/nickrio/coward/common/codec"
	"github.com/nickrio/coward/roles/common/network"
	"github.com/nickrio/coward/roles/common/network/buffer"
	"github.com/nickrio/coward/roles/common/network/messaging"
	"github.com/nickrio/coward/roles/common/network/transporter"
)

type base struct {
	messaging.Messaging

	channelID    byte
	buffer       buffer.Slice
	proc         common.Proccessors
	server       io.ReadWriter
	retryRequest bool
	resetTspConn bool
}

func (b *base) Error(err error) (bool, bool, error) {
	handleErr := err
	tspErr, isTSPErr := err.(transporter.Error)

	if isTSPErr {
		handleErr = tspErr.Raw()
	}

	switch e := handleErr.(type) {
	case codec.Error:
		return false, true, err

	default:
		switch e {
		case io.EOF:
			return b.retryRequest, b.resetTspConn, nil

		case network.ErrProcRemoteTargetClosed:
			fallthrough
		case network.ErrProcServerInternalError:
			fallthrough
		case network.ErrProcUnsupported:
			fallthrough
		case network.ErrProcInvalid:
			fallthrough
		case network.ErrProcServerRefused:
			fallthrough
		case network.ErrProcTimeout:
			fallthrough
		case network.ErrProcRemoteTargetUnconnectable:
			fallthrough
		case network.ErrProcUnsupportedCommand:
			return false, false, err
		}
	}

	return b.retryRequest, b.resetTspConn, err
}

func (b *base) Close() error {
	return nil
}
