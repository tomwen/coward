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
	"net"
	"time"

	ccommon "github.com/nickrio/coward/common"
	"github.com/nickrio/coward/common/codec"
	"github.com/nickrio/coward/roles/common/network"
	"github.com/nickrio/coward/roles/common/network/buffer"
	"github.com/nickrio/coward/roles/common/network/messaging"
	"github.com/nickrio/coward/roles/common/network/transporter"
	"github.com/nickrio/coward/roles/socks5/common"
)

type base struct {
	messaging.Messaging

	buffer        buffer.Slice
	proc          ccommon.Proccessors
	address       []byte
	server        io.ReadWriter
	client        net.Conn
	delayFeedback func(time.Duration)
	retryRequest  bool
	resetTspConn  bool
}

func (b *base) errorRespond(
	conn net.Conn, buffer []byte, err common.REP) error {
	// Tell client we made the connection, message format:
	//
	// +----+-----+-------+------+----------+----------+
	// |VER | REP |  RSV  | ATYP | BND.ADDR | BND.PORT |
	// +----+-----+-------+------+----------+----------+
	// | 1  |  1  | X'00' |  1   | Variable |    2     |
	// +----+-----+-------+------+----------+----------+
	//
	buffer[0] = common.Version // VER
	buffer[1] = 0              // REP
	buffer[2] = byte(err)      // RSV
	buffer[3] = 1              // ATYP: Make up a Address. IPv4 0.0.0.0:0
	buffer[4] = 0              // BND.ADDR IPv4: 0
	buffer[5] = 0              // BND.ADDR IPv4: 0
	buffer[6] = 0              // BND.ADDR IPv4: 0
	buffer[7] = 0              // BND.ADDR IPv4: 0
	buffer[8] = 0              // BND.PORT: 0
	buffer[9] = 0              // BND.PORT

	wLen, wErr := conn.Write(buffer[:10])

	if wErr != nil {
		return wErr
	}

	if wLen != 10 {
		return ErrFailedToSendAllDataToClient
	}

	return nil
}

func (b *base) Error(err error) (bool, bool, error) {
	handleErr := err
	tspErr, isTSPErr := handleErr.(transporter.Error)

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

		case network.ErrProcServerInternalError:
			fallthrough
		case network.ErrProcUnsupported:
			fallthrough
		case network.ErrProcInvalid:
			b.errorRespond(b.client, b.buffer.Client.ExtendedBuffer,
				common.ErrorGeneralFailure)

			return false, false, err

		case network.ErrProcServerRefused:
			b.errorRespond(b.client, b.buffer.Client.ExtendedBuffer,
				common.ErrorForbidden)

			return false, false, err

		case network.ErrProcTimeout:
			b.errorRespond(b.client, b.buffer.Client.ExtendedBuffer,
				common.ErrorHostUnreachable)

			return false, false, err

		case network.ErrProcRemoteTargetUnconnectable:
			b.errorRespond(b.client, b.buffer.Client.ExtendedBuffer,
				common.ErrorConnectionRefused)

			return false, false, err

		case network.ErrProcUnsupportedCommand:
			b.errorRespond(b.client, b.buffer.Client.ExtendedBuffer,
				common.ErrorCommandNotSupported)

			return false, false, err
		}
	}

	// Little rules about whether or not the tsp connection must be
	// resetted:
	//  ANY error that happened before letting server enters relay
	//  status, MUST cause the transporter connection been resetted
	return b.retryRequest, b.resetTspConn, err
}

func (b *base) Close() error {
	return nil
}
