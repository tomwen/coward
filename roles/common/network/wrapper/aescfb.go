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

package wrapper

import (
	"net"
	"time"

	"github.com/nickrio/coward/common/streamer/aes_cfb"
	"github.com/nickrio/coward/roles/common/network"
	"github.com/nickrio/coward/roles/common/network/conn"
	"github.com/nickrio/coward/roles/common/network/transporter/common"
)

// AESCFB128 returns a AES-CFB-128-HMAC Data wrapper
func AESCFB128() Wrapper {
	return Wrapper{
		Name:    "aes-cfb-128-hmac",
		Wrapper: AESCFB128Wrapper,
	}
}

// AESCFB256 returns a AES-CFB-256-HMAC Data wrapper
func AESCFB256() Wrapper {
	return Wrapper{
		Name:    "aes-cfb-256-hmac",
		Wrapper: AESCFB256Wrapper,
	}
}

// AESCFB128Wrapper returns an AES-CFB-128 conn wrapper
func AESCFB128Wrapper(key []byte) common.ConnWrapper {
	timedKey := network.TimedKey(key)

	return func(raw net.Conn) (net.Conn, error) {
		current := time.Now()

		sharedKey, keyErr := timedKey.Get(current, 16)

		if keyErr != nil {
			return nil, keyErr
		}

		cipher, cipherErr := aesCFB.New(sharedKey)

		if cipherErr != nil {
			return nil, cipherErr
		}

		return conn.NewEncoded(raw, cipher), nil
	}
}

// AESCFB256Wrapper returns an AES-CFB-256 conn wrapper
func AESCFB256Wrapper(key []byte) common.ConnWrapper {
	timedKey := network.TimedKey(key)

	return func(raw net.Conn) (net.Conn, error) {
		current := time.Now()

		sharedKey, keyErr := timedKey.Get(current, 32)

		if keyErr != nil {
			return nil, keyErr
		}

		cipher, cipherErr := aesCFB.New(sharedKey)

		if cipherErr != nil {
			return nil, cipherErr
		}

		return conn.NewEncoded(raw, cipher), nil
	}
}
