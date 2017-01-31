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

import (
	"math"
	"testing"
)

func TestEncodableUint16DeEncode(t *testing.T) {
	e := EncodableUint16(16)
	e2 := EncodableUint16(0)

	encodedBytes := make([]byte, 2)

	encodingErr := e.EncodeBytes(encodedBytes)

	if encodingErr != nil {
		t.Error("Failed to encode", e, "due to:", encodingErr)

		return
	}

	if string(encodedBytes) != string([]byte{0, 16}) {
		t.Error("Failed to encode data to expected one. expecting",
			[]byte{0, 16}, "got", encodedBytes)

		return
	}

	decodeErr := e2.DecodeBytes(encodedBytes)

	if decodeErr != nil {
		t.Error("Failed to decode", encodedBytes, "due to error:", decodeErr)

		return
	}

	if e2 != e {
		t.Error("Encode or Decode failed, expected", e, "got", e2)

		return
	}
}

func TestEncodableUint16DeEncode2(t *testing.T) {
	e := EncodableUint16(math.MaxUint16)
	e2 := EncodableUint16(0)

	encodedBytes := make([]byte, 2)

	encodingErr := e.EncodeBytes(encodedBytes)

	if encodingErr != nil {
		t.Error("Failed to encode", e, "due to:", encodingErr)

		return
	}

	decodeErr := e2.DecodeBytes(encodedBytes)

	if decodeErr != nil {
		t.Error("Failed to decode", encodedBytes, "due to error:", decodeErr)

		return
	}

	if e2 != e {
		t.Error("Encode or Decode failed, expected", e, "got", e2)

		return
	}
}

func TestEncodableUint16EncodeInvalidLength(t *testing.T) {
	e := EncodableUint16(math.MaxUint16)

	encodedBytes := make([]byte, 1)

	encodingErr := e.EncodeBytes(encodedBytes)

	if encodingErr != ErrInvalidByteLengthForUInt16 {
		t.Error("Failed to expecting error ErrInvalidByteLengthForUInt16, got",
			encodingErr)
	}
}

func TestEncodableUint16DecodeInvalidLength(t *testing.T) {
	e := EncodableUint16(math.MaxUint16)

	encodedBytes := make([]byte, 1)

	encodingErr := e.DecodeBytes(encodedBytes)

	if encodingErr != ErrInvalidByteLengthForUInt16 {
		t.Error("Failed to expecting error ErrInvalidByteLengthForUInt16, got",
			encodingErr)
	}
}
