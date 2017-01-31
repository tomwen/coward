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

import "errors"

const (
	// Version is the version of current socks server (which is always 5)
	Version byte = 5
)

var (
	// ErrUnsupportedSocksVersion is throwed when client is not version 5
	ErrUnsupportedSocksVersion = errors.New(
		"Unsupported Socks server version")

	// ErrSocks5AddressTooLong is throwed when client is not version 5
	ErrSocks5AddressTooLong = errors.New(
		"Socks5 address is too long")

	// ErrFailedToReadHandshakeHead is throwed when server is failed to read
	// all head bytes (first 2 bytes)
	ErrFailedToReadHandshakeHead = errors.New(
		"Failed to read all handshake head bytes")

	// ErrNoAuthMethodProvided is throwed when client is provided no auth
	// methods
	ErrNoAuthMethodProvided = errors.New(
		"Client doen't provide auth methods")

	// ErrFailedToReadAllAuthMethods is throwed when server is failed to read
	// all auth method bytes (first 2 bytes)
	ErrFailedToReadAllAuthMethods = errors.New(
		"Failed to read all auth method bytes")

	// ErrAuthFailed is throwed when
	ErrAuthFailed = errors.New(
		"Login authentication has failed")

	// ErrUnsupportedAuthMethod is throwed when auth method is not supported
	// by server
	ErrUnsupportedAuthMethod = errors.New(
		"Unsupported auth method")

	// ErrFailedToReadRequestHead is throwed when server is failed to read
	// all head bytes (first 2 bytes)
	ErrFailedToReadRequestHead = errors.New(
		"Failed to read all request head bytes")

	// ErrFailedToReadAuthBytes is throwed when server is failed to read
	// all auth request bytes
	ErrFailedToReadAuthBytes = errors.New(
		"Failed to read all auth request bytes")

	// ErrInvalidAuthCredentialProvided is throwed when client is provided
	// an invalid auth credential
	ErrInvalidAuthCredentialProvided = errors.New(
		"Client didn't provide an valid auth methods")
)

// REP is the REP in Socks 5 server respond
type REP byte

// REPs
const (
	ErrorSucceeded           REP = 0x00
	ErrorGeneralFailure      REP = 0x01
	ErrorForbidden           REP = 0x02
	ErrorNetworkUnreachable  REP = 0x03
	ErrorHostUnreachable     REP = 0x04
	ErrorConnectionRefused   REP = 0x05
	ErrorTTLExpired          REP = 0x06
	ErrorCommandNotSupported REP = 0x07
	ErrorAddressNotSupported REP = 0x08
)
