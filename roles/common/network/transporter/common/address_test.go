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

import "testing"

func TestAddressGet(t *testing.T) {
	a := NewAddress("localhost", 65535)

	addr, addrErr := a.Get()

	if addrErr != nil {
		t.Error("Failed to get Address due to error:", addrErr)

		return
	}

	if addr != "127.0.0.1:65535" && addr != "[::1]:65535" {
		t.Error("Failed to resolve address to expected address. "+
			"Expecting %s, got %s", "127.0.0.1:65535 or [::1]:65535", addr)

		return
	}
}
