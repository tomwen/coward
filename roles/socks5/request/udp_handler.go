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
	"errors"
	"net"

	"github.com/nickrio/coward/roles/common/network/address"
	"github.com/nickrio/coward/roles/common/network/conn"
	"github.com/nickrio/coward/roles/common/network/relay"
	"github.com/nickrio/coward/roles/socks5/common"
)

// UDP handler errors
var (
	ErrInvalidUDPDataPacketLength = errors.New(
		"Invalid data packet length")

	ErrInvalidUDPDataLength = errors.New(
		"Invalid data length")

	ErrUDPFragmentUnsupported = errors.New(
		"Fragment is unsupported")

	ErrInvalidUDPAddressLength = errors.New(
		"Invalid UDP Address length")

	ErrUnknownUDPAddressType = errors.New(
		"Unknown UDP address type")

	ErrUDPPacketSourceUnallowed = errors.New(
		"Unknown UDP source address is not allowed")
)

type udpHandler struct {
	quitter    relay.SignalChan
	onReady    func() error
	clientAddr *net.UDPAddr
	clientIP   net.IP
	addresser  *common.Address
}

func (u *udpHandler) Ready() error {
	return u.onReady()
}

func (u *udpHandler) Quitter() relay.SignalChan {
	return u.quitter
}

// Send read data from internal source and send it to client
func (u *udpHandler) Send(
	c conn.UDPReadWriter, data []byte, writeBuf []byte) (int, error) {
	// Data format
	//
	// +---------------------+-----------------+
	// | ADDRESS INFORMATION |       DATA      |
	// +---------------------+-----------------+
	// |                     |                 |
	// +---------------------+-----------------+

	addrData := address.Address{}

	atype, addr, port, offset, decodeErr := addrData.Unpack(data)

	if decodeErr != nil {
		return 0, decodeErr
	}

	// Convert to
	//
	// +-----+------+------+----------+----------+----------+
	// | RSV | FRAG | ATYP | DST.ADDR | DST.PORT |   DATA   |
	// +-----+------+------+----------+----------+----------+
	// |  2  |  1   |  1   | Variable |    2     | Variable |
	// +-----+------+------+----------+----------+----------+
	//
	// And remember, it's UDP, all data must be send at once

	writeBuf[0] = 0 // RSV
	writeBuf[1] = 0 //
	writeBuf[2] = 0 // FRAG

	// Build ATYP, DST.ADDR and DST.PORT
	var udpAddrErr error
	udpAddrLen := 0

	switch atype {
	case address.IPv4:
		udpAddrLen, udpAddrErr = u.addresser.Pack(
			common.IPv4, addr, port, writeBuf[3:])

	case address.IPv6:
		udpAddrLen, udpAddrErr = u.addresser.Pack(
			common.IPv6, addr, port, writeBuf[3:])

	default:
		return 0, ErrUnknownUDPAddressType
	}

	if udpAddrErr != nil {
		return 0, udpAddrErr
	}

	dataDataLen := len(data[offset:])
	end := 3 + udpAddrLen + dataDataLen

	copy(writeBuf[3+udpAddrLen:end], data[offset:offset+dataDataLen])

	return c.WriteToUDP(writeBuf[:end], u.clientAddr)
}

// Receive receives data from client and convert received data into
// internal format
func (u *udpHandler) Receive(
	c conn.UDPReadWriter, result []byte, readBuf []byte) (int, error) {
	// readBuf: RSV + FRAG + ATYP + DST.ADDR + DST.PORT + DATA = 10 + DATA
	// result:  ADDRESS INFORMATION + DATA = 7 + DATA
	// We must read same amount of bytes of len(result) so data in the
	// readBuf can be fit in to the result buffer
	rLen, rAddr, rErr := c.ReadFromUDP(readBuf)

	if rErr != nil {
		return 0, rErr
	}

	if u.clientAddr == nil && u.clientIP.Equal(rAddr.IP) {
		u.clientAddr = rAddr
	}

	if u.clientAddr == nil ||
		!u.clientAddr.IP.Equal(rAddr.IP) || u.clientAddr.Port != rAddr.Port {
		return 0, ErrUDPPacketSourceUnallowed
	}

	// Expected packet format
	//
	// +-----+------+------+----------+----------+----------+
	// | RSV | FRAG | ATYP | DST.ADDR | DST.PORT |   DATA   |
	// +-----+------+------+----------+----------+----------+
	// |  2  |  1   |  1   | Variable |    2     | Variable |
	// +-----+------+------+----------+----------+----------+

	if rLen < 4 {
		return 0, ErrInvalidUDPDataPacketLength
	}

	// FRAGment is unsupported
	if readBuf[2] != 0 {
		return 0, ErrUDPFragmentUnsupported
	}

	atype, addr, port, offset, err := u.addresser.Unpack(
		readBuf[3:rLen])

	if err != nil {
		return 0, err
	}

	// Convert to
	//
	// +---------------------+-----------------+
	// | ADDRESS INFORMATION |       DATA      |
	// +---------------------+-----------------+
	// |                     |                 |
	// +---------------------+-----------------+

	addrData := address.Address{}

	var addrPackErr error
	var addrPackLen int

	switch atype {
	case common.IPv4:
		addrPackLen, addrPackErr = addrData.Pack(
			address.IPv4, addr, port, result)

	case common.IPv6:
		addrPackLen, addrPackErr = addrData.Pack(
			address.IPv6, addr, port, result)

	case common.Domain:
		addrPackLen, addrPackErr = addrData.Pack(
			address.Domain, addr, port, result)

	default:
		return 0, ErrUnknownUDPAddressType
	}

	if addrPackErr != nil {
		return 0, addrPackErr
	}

	dataDataLen := len(readBuf[offset+3 : rLen])
	end := addrPackLen + dataDataLen

	copy(result[addrPackLen:end], readBuf[offset+3:rLen])

	return end, nil
}
