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

func GetAuther(userVerifier common.AutherUserVerifier) common.Auther {
	a := common.Auther{}

	if userVerifier != nil {
		a.Register(common.PassAuth, func(rw io.ReadWriter, buf []byte) error {
			// Auth select message format:
			//
			// +----+--------+
			// |VER | METHOD |
			// +----+--------+
			// | 1  |   1    |
			// +----+--------+
			//
			//
			// Password Auth message format:
			//
			// +----+------+----------+------+----------+
			// |VER | ULEN |  UNAME   | PLEN |  PASSWD  |
			// +----+------+----------+------+----------+
			// | 1  |  1   | 1 to 255 |  1   | 1 to 255 |
			// +----+------+----------+------+----------+
			//
			//
			// Result message format:
			//
			// +----+--------+
			// |VER | STATUS |
			// +----+--------+
			// | 1  |   1    |
			// +----+--------+

			buf[0] = common.Version
			buf[1] = byte(common.PassAuth)

			_, wErr := rw.Write(buf[:2])

			if wErr != nil {
				return wErr
			}

			// Read message head
			_, rErr := io.ReadFull(rw, buf[:2])

			if rErr != nil {
				return common.ErrFailedToReadAuthBytes
			}

			negoVersion := buf[0]

			if buf[1] <= 0 {
				return common.ErrInvalidAuthCredentialProvided
			}

			// Read User name
			usernameLength := buf[1]

			_, rErr = io.ReadFull(rw, buf[:usernameLength+1])

			if rErr != nil {
				return common.ErrFailedToReadAuthBytes
			}

			username := string(buf[:usernameLength])

			if buf[usernameLength] <= 0 {
				return common.ErrInvalidAuthCredentialProvided
			}

			// Read User password
			passwordLength := buf[usernameLength]

			_, rErr = io.ReadFull(rw, buf[:passwordLength])

			if rErr != nil {
				return common.ErrFailedToReadAuthBytes
			}

			password := string(buf[:passwordLength])

			authErr := userVerifier(username, password)

			buf[0] = negoVersion

			if authErr != nil {
				buf[1] = 1

				_, wErr = rw.Write(buf[:2])

				if wErr != nil {
					return wErr
				}

				return authErr
			}

			buf[1] = 0

			_, wErr = rw.Write(buf[:2])

			if wErr != nil {
				return wErr
			}

			return nil
		})

		return a
	}

	a.Register(common.NoAuth, func(rw io.ReadWriter, buf []byte) error {
		// Result message format:
		//
		// +----+--------+
		// |VER | STATUS |
		// +----+--------+
		// | 1  |   1    |
		// +----+--------+

		buf[0] = common.Version
		buf[1] = byte(common.NoAuth)

		_, wErr := rw.Write(buf[:2])

		if wErr != nil {
			return wErr
		}

		return nil
	})

	return a
}
