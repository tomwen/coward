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

package conn

import (
	"math/rand"
	"net"
	"time"
)

const (
	// minChaoticWriteLen: We only chaoticly write when data length it
	// longer than this value
	minChaoticWriteLen = 1024
)

// chaotic
type chaotic struct {
	net.Conn

	rand *rand.Rand
}

// NewChaotic creates a new chaotic CONN
func NewChaotic(c net.Conn) net.Conn {
	return &chaotic{
		Conn: c,
		rand: rand.New(rand.NewSource(time.Now().Unix())),
	}
}

// Write write to conn
func (c *chaotic) Write(b []byte) (int, error) {
	var wErr error
	var wLen int
	var curPos int
	var curLen int

	remainLen := len(b)
	totalWLen := 0

	for {
		if minChaoticWriteLen >= remainLen {
			wLen, wErr = c.Conn.Write(b[curPos : curPos+remainLen])

			if wErr != nil {
				break
			}

			totalWLen += wLen

			break
		}

		curLen = minChaoticWriteLen + c.rand.Intn(remainLen-minChaoticWriteLen)

		wLen, wErr = c.Conn.Write(b[curPos : curPos+curLen])

		if wErr != nil {
			break
		}

		curPos = curPos + curLen
		remainLen -= curLen
		totalWLen += wLen
	}

	return totalWLen, wErr
}
