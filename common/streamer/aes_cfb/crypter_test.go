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

package aesCFB

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"testing"

	"github.com/nickrio/coward/common/codec"
	"github.com/nickrio/coward/common/types"
)

func testMustGetCrypter(key []byte) *crypter {
	rand, randErr := types.NewRandom(32)

	if randErr != nil {
		panic(randErr)
	}

	return &crypter{
		padding: padding{
			random:     rand,
			cachedByte: [2]byte{},
			max:        16,
		},
		readMac:     hmac.New(sha256.New, key),
		writeMac:    hmac.New(sha256.New, key),
		data:        [codec.BufferSize]byte{},
		sizeEnBytes: [SizeSize]byte{},
		sizeDeBytes: [SizeSize]byte{},
		metaBytes:   [MetaSize]byte{},
	}
}

func TestCrypterHMAC(t *testing.T) {
	c := testMustGetCrypter([]byte("HMAC16BYTELONGKY"))

	hmac1, hmac1Err := c.hmac(c.readMac, []byte("HELLO WORLD"))

	if hmac1Err != nil {
		t.Error("Failed to calecute HMAC due to error:", hmac1Err)

		return
	}

	if string(hmac1) != string(
		[]byte{64, 7, 39, 231, 83, 207, 251, 154, 24, 93}) {
		t.Errorf("HMAC2 is not expected %v, got %v",
			[]byte{64, 7, 39, 231, 83, 207, 251, 154, 24, 93},
			hmac1)

		return
	}

	c2 := testMustGetCrypter([]byte("HMAC16BYTELONGKI"))

	hmac2, hmac2Err := c2.hmac(c2.readMac, []byte("HELLO WORLD"))

	if hmac2Err != nil {
		t.Error("Failed to calecute HMAC due to error:", hmac1Err)

		return
	}

	if string(hmac2) != string(
		[]byte{109, 99, 121, 6, 213, 88, 252, 190, 224, 247}) {
		t.Errorf("HMAC2 is not expected %v, got %v",
			[]byte{109, 99, 121, 6, 213, 88, 252, 190, 224, 247},
			hmac1)

		return
	}
}

func TestCrypterHMACEqual(t *testing.T) {
	c := testMustGetCrypter([]byte("HMAC16BYTELONGKY"))

	hmac1, hmac1Err := c.hmac(c.readMac, []byte("HELLO WORLD"))

	if hmac1Err != nil {
		t.Error("Failed to calecute HMAC due to error:", hmac1Err)

		return
	}

	if string(hmac1) != string(
		[]byte{64, 7, 39, 231, 83, 207, 251, 154, 24, 93}) {
		t.Errorf("HMAC2 is not expected %v, got %v",
			[]byte{64, 7, 39, 231, 83, 207, 251, 154, 24, 93},
			hmac1)

		return
	}

	c2 := testMustGetCrypter([]byte("HMAC16BYTELONGKI"))

	hmac2, hmac2Err := c2.hmac(c2.readMac, []byte("HELLO WORLD"))

	if hmac2Err != nil {
		t.Error("Failed to calecute HMAC due to error:", hmac1Err)

		return
	}

	if string(hmac2) != string(
		[]byte{109, 99, 121, 6, 213, 88, 252, 190, 224, 247}) {
		t.Errorf("HMAC2 is not expected %v, got %v",
			[]byte{109, 99, 121, 6, 213, 88, 252, 190, 224, 247},
			hmac2)

		return
	}

	c3 := testMustGetCrypter([]byte("HMAC16BYTELONGKI"))

	hmac3, hmac3Err := c3.hmac(c3.readMac, []byte("HELLO WORLD"))

	if hmac3Err != nil {
		t.Error("Failed to calecute HMAC due to error:", hmac1Err)

		return
	}

	if string(hmac3) != string(
		[]byte{109, 99, 121, 6, 213, 88, 252, 190, 224, 247}) {
		t.Errorf("HMAC2 is not expected %v, got %v",
			[]byte{109, 99, 121, 6, 213, 88, 252, 190, 224, 247},
			hmac3)

		return
	}

	if c.hmacEqual(hmac1, hmac2) {
		t.Error("Failed to expecting HMAC1 is not equals to HMAC2")

		return
	}

	if !c.hmacEqual(hmac2, hmac3) {
		t.Error("Failed to expecting HMAC2 is equals to HMAC3")

		return
	}
}

func TestCrypterHead(t *testing.T) {
	b, bErr := New([]byte("0123456789ABCDEF"))

	if bErr != nil {
		t.Error("Can't create AES-CFB Builder due to error:", bErr)

		return
	}

	c, cErr := b.New()

	if cErr != nil {
		t.Error("Can't create AES-CFB crypter due to error:", cErr)

		return
	}

	data := make([]byte, 0, 4096)
	buf := bytes.NewBuffer(data)

	headErr := c.Head(buf)

	if headErr != nil {
		t.Error("Can't write Head due to error:", headErr)

		return
	}

	written := buf.Bytes()

	if len(written) != HeadSize {
		t.Errorf("Invalid Head length, expecting %v, got %v",
			HeadSize, len(written))

		return
	}

	if string(written[:HeadSize]) != string(c.(*crypter).head[:]) {
		t.Errorf("Head data is invalid, expecting %v, got %v",
			c.(*crypter).head, written[:HeadSize])

		return
	}
}

func TestCrypterWrapParse(t *testing.T) {
	c := testMustGetCrypter([]byte("HMAC16BYTELONGKI"))
	b := bytes.NewBuffer(make([]byte, 0, 4096))
	input := make([]byte, 16)

	_, randErr := rand.Read(input)

	if randErr != nil {
		t.Error("Failed to generate random seed for this test")

		return
	}

	cErr := c.Stream(input, b)

	if cErr != nil {
		t.Error("Can't wrap due to error:", cErr)

		return
	}

	cErr = c.Stream(input, b)

	if cErr != nil {
		t.Error("Can't wrap due to error:", cErr)

		return
	}

	result, parseErr := c.Unstream(b)

	if parseErr != nil {
		t.Error("Can't wrap due to error:", parseErr)

		return
	}

	if string(result) != string(input) {
		t.Errorf("Failed to wrap or parse input data, expecting %v, got %v",
			input, result)

		return
	}

	result, parseErr = c.Unstream(b)

	if parseErr != nil {
		t.Error("Can't wrap due to error:", cErr)

		return
	}

	if string(result) != string(input) {
		t.Errorf("Failed to wrap or parse input data, expecting %v, got %v",
			input, result)

		return
	}
}

func TestCrypterStreamUnstream(t *testing.T) {
	b, bErr := New([]byte("0123456789ABCDEF"))

	testData := make([]byte, 128)

	_, randErr := rand.Read(testData)

	if randErr != nil {
		t.Error("Failed to generate random seed for this test")

		return
	}

	if bErr != nil {
		t.Error("Can't create AES-CFB driver for the test due to error:", bErr)

		return
	}

	c, cErr := b.New()

	if cErr != nil {
		t.Error("Can't create AES-CFB chiper due to error:", cErr)

		return
	}

	data := make([]byte, 0, 4096)
	buf := bytes.NewBuffer(data)
	input := make([]byte, len(testData))

	copy(input, testData)

	headErr := c.Head(buf)

	if headErr != nil {
		t.Error("Can't save head due to error:", headErr)

		return
	}

	encodeErr := c.Stream(input, buf)

	if encodeErr != nil {
		t.Error("Can't encode due to error:", encodeErr)

		return
	}

	d, cErr := b.Init(buf)

	if cErr != nil {
		t.Error("Can't create AES-CFB chiper due to error:", cErr)

		return
	}

	decoded, decodeErr := d.Unstream(buf)

	if decodeErr != nil {
		t.Error("Can't decode due to error:", decodeErr)

		return
	}

	if string(decoded) != string(testData) {
		t.Errorf("Can't encode or decode data.\r\nExpecting %v\r\nGot       %v",
			testData, decoded)

		return
	}
}

func BenchmarkCrypterWrapParse(b *testing.B) {
	c := testMustGetCrypter([]byte("HMAC16BYTELONGKI"))
	buf := bytes.NewBuffer(make([]byte, 0, 4096))
	input := make([]byte, 4096)

	_, randErr := rand.Read(input)

	if randErr != nil {
		b.Error("Failed to generate random seed for this test")

		return
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cErr := c.Stream(input, buf)

		if cErr != nil {
			b.Error("Can't wrap due to error:", cErr)

			return
		}

		result, parseErr := c.Unstream(buf)

		if parseErr != nil {
			b.Error("Can't wrap due to error:", cErr)

			return
		}

		if string(result) != string(input) {
			b.Errorf("Failed to wrap or parse input data, expecting %v, got %v",
				input, result)

			return
		}
	}
}
