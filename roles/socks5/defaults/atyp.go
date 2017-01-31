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

package defaults

import (
	"io"

	"github.com/nickrio/coward/roles/socks5/common"
)

// GetATYPStream return a new ATYPE stream Parser
func GetATYPStream() common.ATYP {
	a := common.ATYP{}

	a.Register(common.IPv4,
		func(rw io.ReadWriter, buf []byte) ([]byte, error) {
			_, rErr := io.ReadFull(rw, buf[:4])

			if rErr != nil {
				return nil, rErr
			}

			return buf[:4], nil
		}).Register(common.Domain,
		func(rw io.ReadWriter, buf []byte) ([]byte, error) {
			rLen, rErr := io.ReadFull(rw, buf[:1])

			if rErr != nil {
				return nil, rErr
			}

			domainLength := buf[0]

			rLen, rErr = io.ReadFull(rw, buf[:domainLength])

			if rErr != nil {
				return nil, rErr
			}

			return buf[:rLen], nil
		}).Register(common.IPv6,
		func(rw io.ReadWriter, buf []byte) ([]byte, error) {
			_, rErr := io.ReadFull(rw, buf[:16])

			if rErr != nil {
				return nil, rErr
			}

			return buf[:16], nil
		})

	return a
}

// GetATYPBlock return a new ATYPE block Parser
func GetATYPBlock() common.Address {
	return common.Address{}
}
