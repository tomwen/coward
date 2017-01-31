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
	"sync"
	"time"

	ccommon "github.com/nickrio/coward/common"
	"github.com/nickrio/coward/roles/common/network/messaging"
	"github.com/nickrio/coward/roles/common/network/relay"
	"github.com/nickrio/coward/roles/common/network/transporter"
	"github.com/nickrio/coward/roles/socks5/common"
)

// UDP request errors
var (
	ErrClientConnectionClosed = errors.New(
		"Client connection closed")
)

type udp struct {
	base

	addresser *common.Address
	addrType  common.ATYPE
}

// NewUDPRequest creates a new UDP request
func NewUDPRequest(
	config transporter.HandlerConfig,
	proc ccommon.Proccessors,
	addresser *common.Address,
	client net.Conn,
	targetType common.ATYPE,
	targetAddr []byte,
	targetPort []byte,
	delayFeedback func(time.Duration),
) transporter.Handler {
	return &udp{
		base: base{
			buffer:        config.Buffer,
			proc:          proc,
			address:       append(targetAddr, targetPort...),
			server:        config.Server,
			client:        client,
			delayFeedback: delayFeedback,
			retryRequest:  false,
			resetTspConn:  false,
		},
		addresser: addresser,
		addrType:  targetType,
	}
}

func (u *udp) notifiyBind(addr *net.UDPAddr, buf []byte) error {
	//  +-----+-----+-------+------+----------+----------+
	//  | VER | REP |  RSV  | ATYP | BND.ADDR | BND.PORT |
	//  +-----+-----+-------+------+----------+----------+
	//  |  1  |  1  | X'00' |  1   | Variable |    2     |
	//  +-----+-----+-------+------+----------+----------+
	buf[0] = common.Version // VER
	buf[1] = 0              // REP
	buf[2] = 0              // RSV

	pLen, pErr := u.addresser.PackIP(
		addr.IP, uint16(addr.Port), buf[3:])

	if pErr != nil {
		return pErr
	}

	_, wErr := u.client.Write(buf[:pLen+3])

	if wErr != nil {
		return wErr
	}

	return nil
}

func (u *udp) Handle() error {
	startTime := time.Now()
	wait := sync.WaitGroup{}

	defer func() {
		wait.Wait()

		u.client.Close()
	}()

	// 0, Enable retry and reset
	u.resetTspConn = true
	u.retryRequest = true

	// 1, Get our IP Address
	localHost, _, localAddrErr := net.SplitHostPort(
		u.client.LocalAddr().String())

	if localAddrErr != nil {
		return localAddrErr
	}

	// 2, Create a UDP listener at ephemeral port
	udpAddr, udpAddrErr := net.ResolveUDPAddr("udp", localHost+":0")

	if udpAddrErr != nil {
		return udpAddrErr
	}

	udpListener, udpListenErr := net.ListenUDP("udp", udpAddr)

	if udpListenErr != nil {
		return udpListenErr
	}

	defer udpListener.Close()

	// 4, Ask server to open UDP port
	_, wErr := u.Write(u.server, messaging.RelayUDP,
		nil, u.buffer.Server.ExtendedBuffer)

	if wErr != nil {
		return wErr
	}

	dispErr := u.Dispatch(u.server, u.buffer.Server.Buffer, u.proc)

	if dispErr != nil {
		return dispErr
	}

	u.delayFeedback(time.Now().Sub(startTime))

	// 5, Start to monitoring the client TCP connection, exit relay
	//   as it drops
	relayQuitterChan := make(relay.SignalChan)

	wait.Add(1)

	go func() {
		defer func() {
			select {
			case relayQuitterChan <- ErrClientConnectionClosed:
			default:
			}

			wait.Done()
		}()

		buffer := [1]byte{}

		for {
			_, rErr := u.client.Read(buffer[:])

			if rErr != nil {
				break
			}
		}
	}()

	// 6, Disable request retry
	u.retryRequest = false
	u.resetTspConn = false

	// 7, Start data sync form proxy to server
	return relay.NewUDPRelay(
		&udpHandler{
			quitter: relayQuitterChan,
			onReady: func() error {
				// 3, Notifiy the client that it can send UDP packets there
				replyErr := u.notifiyBind(
					udpListener.LocalAddr().(*net.UDPAddr),
					u.buffer.Client.Buffer)

				if replyErr != nil {
					udpListener.Close()

					return replyErr
				}

				return nil
			},
			clientIP:  u.client.RemoteAddr().(*net.TCPAddr).IP,
			addresser: u.addresser,
		},
		udpListener,
		u.server,
		u.buffer,
		nil,
	).Relay()
}
