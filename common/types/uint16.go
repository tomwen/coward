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

package types

import "errors"

var (
	// ErrInvalidByteLengthForUInt16 indicates input byte length is
	// invalid for an uint16
	ErrInvalidByteLengthForUInt16 = errors.New(
		"Invalid byte length for uint16")
)

const (
	// Uint16ByteSize indicates how many bytes is needed to represent
	// a uint16 data
	Uint16ByteSize = 2 // = 16 / 8
)

// EncodableUint16 is a type that can convert uint16 into or
// from a slice of bytes
type EncodableUint16 uint16

// DecodeBytes Decode bytes form an uint16 data
func (e *EncodableUint16) DecodeBytes(b []byte) error {
	bLen := len(b)

	if bLen != Uint16ByteSize {
		return ErrInvalidByteLengthForUInt16
	}

	*e = EncodableUint16((uint16(b[0]) << 8) | uint16(b[1]))

	return nil
}

// EncodeBytes Decode uint16 to a byte slice
func (e *EncodableUint16) EncodeBytes(buf []byte) error {
	bufLen := len(buf)

	if bufLen != Uint16ByteSize {
		return ErrInvalidByteLengthForUInt16
	}

	buf[0] = byte(*e >> 8)
	buf[1] = byte((*e << 8) >> 8)

	return nil
}
