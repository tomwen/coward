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
	"errors"
	"io"
	"net"

	"github.com/nickrio/coward/roles/common/network/address"
	"github.com/nickrio/coward/roles/common/network/conn"
	"github.com/nickrio/coward/roles/common/network/messaging"
	"github.com/nickrio/coward/roles/common/network/relay"
)

// udpHandle Errors
var (
	ErrUnsupportedUDPAddressType = errors.New(
		"Unsupported UDP address type")

	ErrFailedResolveHost = errors.New(
		"Failed to resolve host")
)

// Ephemeral public port
var (
	udpEphemeralListenAddr, udpEphemeralListenErr = net.ResolveUDPAddr(
		"udp", "0.0.0.0:0")
)

// UDP Connection errors
var (
	ErrInvalidUDPEphemeralPortAddr = errors.New(
		"Invalid UDP ephemeral port address")

	ErrFailedToOpenUDPEphemeralPort = errors.New(
		"Failed to open new UDP ephemeral port")
)

// udpHandle is the UDP relay handler
type udpHandle struct {
	lastIP        net.IP
	lastPort      int
	isValidTarget func(*net.UDPAddr) error
}

// Ready is a callback that will be call when relay is ready to
// work
func (h *udpHandle) Ready() error {
	return nil
}

// Quitter quits the relay if triggered
func (h *udpHandle) Quitter() relay.SignalChan {
	return nil
}

// Send reads data from internal source and send it to remote
// according to instructions contained in data
func (h *udpHandle) Send(
	c conn.UDPReadWriter, data []byte, wBuf []byte) (int, error) {
	// Data:
	//
	// +---------------------+-----------------+
	// | ADDRESS INFORMATION |       DATA      |
	// +---------------------+-----------------+
	// |                     |                 |
	// +---------------------+-----------------+

	// Handle the ADDRESS INFORMATION
	atype, addr, port, offset, err := address.Default.Unpack(data)

	if err != nil {
		return 0, err
	}

	udpAddr := net.UDPAddr{}

	switch atype {
	case address.IPv4:
		fallthrough
	case address.IPv6:
		udpAddr.IP = addr
		udpAddr.Port = int(port)

	case address.Domain:
		ips, lookupErr := net.LookupIP(string(addr))

		if lookupErr != nil || len(ips) < 1 {
			return 0, ErrFailedResolveHost
		}

		udpAddr.IP = ips[0]
		udpAddr.Port = int(port)

	case address.UseLast:
		udpAddr.IP = h.lastIP
		udpAddr.Port = h.lastPort

	default:
		return 0, ErrUnsupportedUDPAddressType
	}

	targetVerifyErr := h.isValidTarget(&udpAddr)

	if targetVerifyErr != nil {
		return 0, targetVerifyErr
	}

	// Handle the data & send
	return c.WriteToUDP(data[offset:], &udpAddr)
}

// Received receives data from remote and convert received data
// to internal format
func (h *udpHandle) Receive(
	c conn.UDPReadWriter, buf []byte, rBuf []byte) (int, error) {
	rLen, rAddr, rErr := c.ReadFromUDP(rBuf)

	if rErr != nil {
		return 0, rErr
	}

	// Data:
	//
	// +---------------------+-----------------+
	// | ADDRESS INFORMATION |       DATA      |
	// +---------------------+-----------------+
	// |                     |                 |
	// +---------------------+-----------------+

	// Handle the ADDRESS INFORMATION
	rAddrEncoded := address.UDP(*rAddr)

	encodeLen, encodeErr := rAddrEncoded.Encode(buf)

	if encodeErr != nil {
		return 0, encodeErr
	}

	// Apply DATA
	copy(buf[encodeLen:], rBuf[:rLen])

	return rLen + encodeLen, nil
}

func (h *handler) udp(
	buffer []byte,
	client io.ReadWriter,
	size uint16,
) error {
	if udpEphemeralListenErr != nil {
		return ErrInvalidUDPEphemeralPortAddr
	}

	// Request an ephemeral UDP port from OS
	udpConn, udpListenerErr := net.ListenUDP(
		"udp", udpEphemeralListenAddr)

	if udpListenerErr != nil {
		return ErrFailedToOpenUDPEphemeralPort
	}

	defer udpConn.Close()

	// Tell client we all set and ready to roll
	_, wErr := h.Write(client, messaging.OK, nil,
		h.buffer.Client.ExtendedBuffer)

	if wErr != nil {
		return ErrFailedSendConnectConfirmSignal
	}

	// Starting relay
	return relay.NewUDPRelay(
		&udpHandle{
			isValidTarget: func(udpAddr *net.UDPAddr) error {
				if udpAddr.IP.IsUnspecified() {
					return ErrZeroAddressIsForbidden
				}

				if udpAddr.IP.IsLoopback() {
					return ErrLoopbackAddressIsForbidden
				}

				if udpAddr.Port <= 0 {
					return ErrZeroPortIsForbidden
				}

				return nil
			},
		},
		udpConn,
		h.client,
		h.buffer,
		h.closeChan,
	).Relay()
}
