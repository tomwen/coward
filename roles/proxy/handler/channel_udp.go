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

	"github.com/nickrio/coward/roles/common/network"
	"github.com/nickrio/coward/roles/common/network/conn"
	"github.com/nickrio/coward/roles/common/network/messaging"
	"github.com/nickrio/coward/roles/common/network/relay"
)

// Channel UDP errors
var (
	ErrChannelUDPPacketSourceNotAllowed = errors.New(
		"Received UDP Packet is sent by an unknown source")
)

type channelUDPHander struct {
	allowedSource *net.UDPAddr
}

func (c *channelUDPHander) Ready() error {
	return nil
}

func (c *channelUDPHander) Quitter() relay.SignalChan {
	return nil
}

func (c *channelUDPHander) Send(
	u conn.UDPReadWriter, data []byte, wBuf []byte) (int, error) {
	return u.WriteToUDP(data, c.allowedSource)
}

func (c *channelUDPHander) Receive(
	u conn.UDPReadWriter, buf []byte, rBuf []byte) (int, error) {
	rLen, rAddr, rErr := u.ReadFromUDP(rBuf)

	if rErr != nil {
		return 0, rErr
	}

	if !c.allowedSource.IP.Equal(rAddr.IP) {
		return 0, ErrChannelUDPPacketSourceNotAllowed
	}

	copy(buf[:rLen], rBuf[:rLen])

	return rLen, nil
}

func (h *handler) channelUDP(
	buffer []byte,
	client io.ReadWriter,
	size uint16,
) error {
	// Channel UDP format:
	//
	// +------------+
	// | Channel ID |
	// +------------+
	// |     1      |
	// +------------+
	//
	// Expected:
	// Channel ID:   1

	// Check if we failed to generate UDP listen address
	if udpEphemeralListenErr != nil {
		return ErrInvalidUDPEphemeralPortAddr
	}

	// Check buffer size
	if size <= 0 {
		return ErrInvalidChannelID
	}

	// Read data for Channel ID
	_, readAddrErr := io.ReadFull(client, buffer[:size])

	if readAddrErr != nil {
		return readAddrErr
	}

	// Yeah, only read the first byte as Channel ID
	ch, chErr := h.channels.Get(buffer[0])

	if chErr != nil || ch.Protocol != network.UDP {
		return ErrRequestingUndefindedChannel
	}

	// Request an ephemeral UDP port from OS
	udpConn, udpListenerErr := net.ListenUDP(
		"udp", udpEphemeralListenAddr)

	if udpListenerErr != nil {
		return ErrFailedToOpenUDPEphemeralPort
	}

	defer udpConn.Close()

	// Parse Channel Target Address to UDP address
	targetAddr, targetAddrResolveErr := net.ResolveUDPAddr("udp", ch.Address)

	if targetAddrResolveErr != nil {
		return targetAddrResolveErr
	}

	// Tell client we are set
	_, wErr := h.Write(client, messaging.OK, nil,
		h.buffer.Client.ExtendedBuffer)

	if wErr != nil {
		return ErrFailedSendConnectConfirmSignal
	}

	// Starting relay
	return relay.NewUDPRelay(
		&channelUDPHander{
			allowedSource: targetAddr,
		},
		udpConn,
		h.client,
		h.buffer,
		h.closeChan,
	).Relay()
}
