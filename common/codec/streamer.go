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

package codec

import "io"

var (
	// ErrStreamOverheadTooLarge throws when Cipher overhead is too large
	ErrStreamOverheadTooLarge = Fail(
		"Stream overhead too large")

	// ErrStreamUnexpectedDataSegmentLength throws when length of readed data
	// is unexpected
	ErrStreamUnexpectedDataSegmentLength = Fail(
		"Unexpected data segment length")
)

// Streamer interface defineded methods that required to build a Wrapper
type Streamer interface {
	New() (Stream, error)
	Init(io.Reader) (Stream, error)

	OverheadSize() int
}

// Stream interface represents an object can be use to wrap and unwrap
// data
type Stream interface {
	Head(writer io.Writer) error

	Stream(input []byte, w io.Writer) error
	Encode(input []byte, w io.Writer) error

	Decode(input []byte) (decoded []byte, decodeErr error)
	Unstream(r io.Reader) (output []byte, err error)
}
