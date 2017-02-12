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

package aescfb

import (
	"bytes"
	"crypto/rand"
	"io"
	"testing"

	"github.com/nickrio/coward/common/codec"
)

func TestBuilder(t *testing.T) {
	data := make([]byte, 0, 4096)
	buf := bytes.NewBuffer(data)
	cipher, cipherErr := New([]byte("1234567890123456"))
	testData := [][]byte{
		[]byte("Test data is here"),
		[]byte("Test data is here again"),
	}

	if cipherErr != nil {
		t.Error("Cipher creation failed:", cipherErr)

		return
	}

	reader := codec.NewReader(cipher, buf)
	writer := codec.NewWriter(cipher, buf)

	for _, data := range testData {
		wLen, wErr := writer.Write(data)

		if wErr != nil {
			t.Error("Writer failed:", wErr)

			return
		}

		if wLen != len(data) {
			t.Error("Writer failed: write incompleted")

			return
		}
	}

	count := 0

	for {
		readBuf := make([]byte, codec.BufferSize*2)

		rLen, rErr := reader.Read(readBuf)

		if rErr != nil {
			t.Error("Can't read due to error:", rErr, rLen)

			return
		}

		if string(testData[count]) != string(readBuf[:rLen]) {
			t.Errorf("Failed to read expected data, expecting %v, got %v",
				testData[count], readBuf[:rLen])

			return
		}

		count++

		if count >= 2 {
			break
		}
	}
}

func BenchmarkBuilderReadWrite(b *testing.B) {
	data := make([]byte, 0, 4096)
	buf := bytes.NewBuffer(data)
	cipher, cipherErr := New([]byte("1234567890123456"))

	if cipherErr != nil {
		b.Error("Cipher creation failed:", cipherErr)

		return
	}

	testData := make([]byte, 4096)

	_, randErr := rand.Read(testData)

	if randErr != nil {
		b.Error("Can't generate random bytes for this benchmark due to error:",
			randErr)

		return
	}

	reader := codec.NewReader(cipher, buf)
	writer := codec.NewWriter(cipher, buf)

	readBuf := make([]byte, codec.BufferSize*2)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		wLen, wErr := writer.Write(testData)

		if wErr != nil {
			b.Error("Writer failed:", wErr)

			return
		}

		if wLen != len(testData) {
			b.Error("Writer failed: write incompleted")

			return
		}

		for {
			rLen, rErr := reader.Read(readBuf)

			if rErr == nil {
				continue
			}

			if rErr == io.EOF {
				break
			}

			b.Error("Can't read due to error:", rErr, rLen)

			return
		}
	}
}
