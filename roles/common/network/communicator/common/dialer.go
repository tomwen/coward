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

package common

import (
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/nickrio/coward/common/locked"
)

// Dialer is a net.Dial wrapper which will cache target IP address
type Dialer interface {
	Dial(dialType string, dialTimeout time.Duration) (net.Conn, error)
}

type dialer struct {
	defaultHost    string
	resolvedHostIP net.IP
	port           uint16
	tryRenew       locked.Boolean
	lock           sync.Mutex
}

// NewDialer creates a new Dialer
func NewDialer(defaultHost string, port uint16) Dialer {
	return &dialer{
		defaultHost:    defaultHost,
		resolvedHostIP: nil,
		port:           port,
		tryRenew:       locked.NewBool(false),
		lock:           sync.Mutex{},
	}
}

// Dial dial to the remote target
func (d *dialer) Dial(
	dialType string, dialTimeout time.Duration) (net.Conn, error) {
	var targetAddr string
	var updateResolvedIP bool

	if d.resolvedHostIP == nil || d.tryRenew.Get() {
		targetAddr = d.defaultHost
		updateResolvedIP = true
	} else {
		targetAddr = d.resolvedHostIP.String()
	}

	d.tryRenew.Set(false)

	hostAddr := net.JoinHostPort(
		targetAddr, strconv.FormatUint(uint64(d.port), 10))

	conn, dialErr := net.DialTimeout(dialType, hostAddr, dialTimeout)

	if dialErr != nil {
		d.tryRenew.Set(true)

		return nil, dialErr
	}

	if updateResolvedIP {
		d.lock.Lock()
		defer d.lock.Unlock()

		connIP, _, connSpErr := net.SplitHostPort(
			conn.RemoteAddr().String())

		if connSpErr == nil {
			d.resolvedHostIP = net.ParseIP(connIP)
		}
	}

	return conn, dialErr
}
