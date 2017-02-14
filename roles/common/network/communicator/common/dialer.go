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
// The reason for Dialer to exist is for some network, sometimes it's
// hard to always get in touch with some DNS servers to get a host name
// resolved.
// Consider the host IP is usually stay the same, we could keep connecting
// it using the remote IP address resolved by pervious dialer.
// + We also implemented a method to renew the resolved IP address when
// needed: Try to renew the resolved IP every time when a dial has failed.
// but do not erase the old IP if resolver has failed.
type Dialer interface {
	Dial(dialType string, dialTimeout time.Duration) (net.Conn, error)
}

// dialer implements Dialer
type dialer struct {
	defaultHost    string
	resolvedHostIP net.IP
	port           uint16
	tryRenew       locked.Boolean
	lock           sync.RWMutex
}

// NewDialer creates a new Dialer
func NewDialer(defaultHost string, port uint16) Dialer {
	return &dialer{
		defaultHost:    defaultHost,
		resolvedHostIP: nil,
		port:           port,
		tryRenew:       locked.NewBool(false),
		lock:           sync.RWMutex{},
	}
}

// updateIPResolved updates resolved IP address from net.Conn
func (d *dialer) updateIPResolved(conn net.Conn) {
	connIP, _, connSpErr := net.SplitHostPort(
		conn.RemoteAddr().String())

	if connSpErr != nil {
		return
	}

	resolvedIP := net.ParseIP(connIP)

	if resolvedIP == nil {
		return
	}

	d.lock.Lock()
	defer d.lock.Unlock()

	d.resolvedHostIP = resolvedIP
}

// getResolvedIP returns resolved IP from cache
func (d *dialer) getResolvedIP() string {
	d.lock.RLock()
	defer d.lock.RUnlock()

	return d.resolvedHostIP.String()
}

// Dial dial to the remote target
func (d *dialer) Dial(
	dialType string, dialTimeout time.Duration) (net.Conn, error) {
	var targetAddr string
	var updateResolvedIP bool

	// We will set tryRenew to false no matter if we actually got a
	// new IP address.
	// This is because some time network itself can cause unstable
	// connection which triggers tryRenew, but that not meaning the
	// IP address of the target host has changed.
	// If we clear the resolvedHostIP for every each failed connection
	// that will be a waste of work.

	resolvedHostIP := d.getResolvedIP()

	if d.tryRenew.GetSet(false) || len(resolvedHostIP) <= 0 {
		targetAddr = d.defaultHost
		updateResolvedIP = true
	} else {
		targetAddr = resolvedHostIP
	}

	hostAddr := net.JoinHostPort(
		targetAddr, strconv.FormatUint(uint64(d.port), 10))

	conn, dialErr := net.DialTimeout(dialType, hostAddr, dialTimeout)

	if dialErr != nil {
		d.tryRenew.Set(true)

		return nil, dialErr
	}

	if updateResolvedIP {
		d.updateIPResolved(conn)
	}

	return conn, dialErr
}
