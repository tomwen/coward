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

import "io"

// AuthMethod is the Auth method
type AuthMethod byte

// AuthHandler is the auth handler
type AuthHandler func(io.ReadWriter, []byte) error

// AutherUserVerifier is a function will verify user name and password
type AutherUserVerifier func(string, string) error

// Known auth types
const (
	NoAuth     AuthMethod = 0x0
	PassAuth   AuthMethod = 0x2
	Unauthable AuthMethod = 0xff
	Max        AuthMethod = PassAuth // Max auth method ID
)

// Auther is the Auth hander
type Auther struct {
	auther [Max + 1]AuthHandler
}

// Register registers a new auth method
func (a *Auther) Register(methodName AuthMethod, method AuthHandler) {
	a.auther[methodName] = method
}

// Has checks if the auth method is registered
func (a *Auther) Has(methodName AuthMethod) bool {
	if methodName > Max {
		return false
	}

	if a.auther[methodName] == nil {
		return false
	}

	return true
}

// Auth do the auth according to the auth method
func (a *Auther) Auth(
	methodName AuthMethod, rw io.ReadWriter, buf []byte) error {
	return a.auther[methodName](rw, buf)
}
