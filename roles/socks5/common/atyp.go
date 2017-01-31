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
	"errors"
	"io"
)

// ATYPE is the address type
type ATYPE byte

var (
	// ErrUnsupportedAddressType is throwed when address type
	// is unsupported
	ErrUnsupportedAddressType = errors.New(
		"Unsupported address type")
)

// ATYPEs
const (
	Unknown ATYPE = 0x00
	IPv4    ATYPE = 0x01
	Domain  ATYPE = 0x03
	IPv6    ATYPE = 0x04
)

// ATYPS is the size of ATYP
const ATYPS = 5

// ATYPER is the type parser
type ATYPER func(io.ReadWriter, []byte) ([]byte, error)

// ATYP is address parser
type ATYP struct {
	atypes [ATYPS]ATYPER
}

// Register registers a new ATYPE parser
func (a *ATYP) Register(atype ATYPE, h ATYPER) *ATYP {
	a.atypes[atype] = h

	return a
}

// Run parse the data according to the ATYPE
func (a *ATYP) Run(
	atype ATYPE, rw io.ReadWriter, buf []byte) (ATYPE, []byte, error) {
	if atype >= ATYPS {
		return atype, nil, ErrUnsupportedAddressType
	}

	if a.atypes[atype] == nil {
		return atype, nil, ErrUnsupportedAddressType
	}

	result, err := a.atypes[atype](rw, buf)

	return atype, result, err
}
