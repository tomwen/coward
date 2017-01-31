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

package buffer

import (
	"bytes"
	"io"
	"testing"
)

type dummyByteBuffer struct {
	r      io.ReadWriter
	rCount int
	wCount int
}

func (d *dummyByteBuffer) Read(b []byte) (int, error) {
	d.rCount++

	return d.r.Read(b)
}

func (d *dummyByteBuffer) Write(b []byte) (int, error) {
	d.wCount++

	return d.r.Write(b)
}

func TestBufferRead1024BytesOneByOneThenEOF(t *testing.T) {
	testData := testGetRandomBytes(testBufferSize)
	source := &dummyByteBuffer{
		r:      bytes.NewBuffer(nil),
		rCount: 0,
		wCount: 0,
	}

	source.Write(testData)

	bufferBuffer := NewBuffer(testBufferSize, source)

	readBuf := make([]byte, 1024)

	// First read:
	// Expecting to read 1024 bytes data from the source
	// which is the first 1024 bytes of our testData.
	// Source will be readed 1 time
	bufferReadLen, bufferReadErr := bufferBuffer.Read(readBuf)

	if bufferReadErr != nil {
		t.Error("Can't read buffer due to error:", bufferReadErr)

		return
	}

	if bufferReadLen != len(readBuf) {
		t.Errorf("Can't read all data. Expected %d, readed %d",
			len(readBuf), bufferReadLen)

		return
	}

	if string(readBuf) != string(testData[:1024]) {
		t.Errorf("Failed to read correct data. Expecting %d, got %d",
			testData[:1024], readBuf)

		return
	}

	if source.rCount != 1 {
		t.Errorf("We should read from source 1 time, got %d",
			source.rCount)

		return
	}

	// Second read:
	// Expecting to read the next 1024 bytes from source
	// which is the second 1024 bytes of our testData.
	// Source will not be readed as we already has enough data
	// in buffer
	bufferReadLen, bufferReadErr = bufferBuffer.Read(readBuf)

	if bufferReadErr != nil {
		t.Error("Can't read buffer due to error:", bufferReadErr)

		return
	}

	if bufferReadLen != len(readBuf) {
		t.Errorf("Can't read all data. Expected %d, readed %d",
			len(readBuf), bufferReadLen)

		return
	}

	if string(readBuf) != string(testData[1024:2048]) {
		t.Errorf("Failed to read correct data. Expecting %d, got %d",
			testData[1024:2048], readBuf)

		return
	}

	if source.rCount != 1 {
		t.Errorf("We should read from source 1 time, got %d",
			source.rCount)

		return
	}

	// Third read:
	// Expecting to read the next 1024 bytes from source
	// which is the third 1024 bytes of our testData.
	// Source will not be readed as we already has enough data
	// in buffer
	bufferReadLen, bufferReadErr = bufferBuffer.Read(readBuf)

	if bufferReadErr != nil {
		t.Error("Can't read buffer due to error:", bufferReadErr)

		return
	}

	if bufferReadLen != len(readBuf) {
		t.Errorf("Can't read all data. Expected %d, readed %d",
			len(readBuf), bufferReadLen)

		return
	}

	if string(readBuf) != string(testData[2048:3072]) {
		t.Errorf("Failed to read correct data. Expecting %d, got %d",
			testData[2048:3072], readBuf)

		return
	}

	if source.rCount != 1 {
		t.Errorf("We should read from source 1 time, got %d",
			source.rCount)

		return
	}

	// Last read:
	// Expecting to read the next 1024 bytes from source
	// which is the last 1024 bytes of our testData.
	// Source will not be readed as we already has enough data
	// in buffer
	bufferReadLen, bufferReadErr = bufferBuffer.Read(readBuf)

	if bufferReadErr != nil {
		t.Error("Can't read buffer due to error:", bufferReadErr)

		return
	}

	if bufferReadLen != len(readBuf) {
		t.Errorf("Can't read all data. Expected %d, readed %d",
			len(readBuf), bufferReadLen)

		return
	}

	if string(readBuf) != string(testData[3072:]) {
		t.Errorf("Failed to read correct data. Expecting %d, got %d",
			testData[3072:], readBuf)

		return
	}

	if source.rCount != 1 {
		t.Errorf("We should read from source 1 time, got %d",
			source.rCount)

		return
	}

	// This will lead us to a EOF
	bufferReadLen, bufferReadErr = bufferBuffer.Read(readBuf)

	if bufferReadErr != io.EOF {
		t.Error("Failed to get an EOF error, got", bufferReadErr)

		return
	}

	if bufferReadLen != 0 {
		t.Error("EOF error should means 0 bytes been readed, instead we have",
			bufferReadLen)

		return
	}

	if source.rCount != 2 {
		t.Errorf("Here should be 2 reads, one succeed, one failed, "+
			"instead we got %d", source.rCount)

		return
	}
}

func TestBufferRead4096PlusOneBytes(t *testing.T) {
	testData := testGetRandomBytes(testBufferSize * 3)
	source := &dummyByteBuffer{
		r:      bytes.NewBuffer(nil),
		rCount: 0,
		wCount: 0,
	}

	source.Write(testData)

	bufferBuffer := NewBuffer(testBufferSize, source)

	readBuf := make([]byte, testBufferSize+1)

	// First read:
	// Expecting to read 4097 bytes data from the source
	// which is the first 4097 bytes of our testData.
	// Since we have 4096 bytes buffer and wants to read 4097
	// bytes data, the source will be readed 2 times
	bufferReadLen, bufferReadErr := bufferBuffer.Read(readBuf)

	if bufferReadErr != nil {
		t.Error("Can't read buffer due to error:", bufferReadErr)

		return
	}

	if bufferReadLen != len(readBuf) {
		t.Errorf("Can't read all data. Expected %d, readed %d",
			len(readBuf), bufferReadLen)

		return
	}

	if string(readBuf) != string(testData[:testBufferSize+1]) {
		t.Errorf("Failed to read correct data. Expecting %d, got %d",
			testData[:testBufferSize+1], readBuf)

		return
	}

	if bufferBuffer.Len() != 4095 {
		t.Errorf("There should be 4095 bytes remained in buffer, got %d",
			bufferBuffer.Len())

		return
	}

	if source.rCount != 2 {
		t.Errorf("We should read from source 2 times, got %d",
			source.rCount)

		return
	}

	// Second read:
	// Expecting to read 4097 bytes data from the source
	// which is the second 4097 bytes of our testData.
	// Since we have 4095 bytes buffer lefted by pervious
	// step and wants to read 2 bytes from the source so
	// the source will be readed 1 time.
	//  Source read should be like:
	//      1. Take copy all existing data to the readBuf
	//         which makes the first 4095 bytes
	//      2. Read source for another 4096 bytes
	//      3. Copy first two bytes from that 4096 bytes to
	//         readBuf and we're done. 4094 bytes now remain
	//         in the buffer
	bufferReadLen, bufferReadErr = bufferBuffer.Read(readBuf)

	if bufferReadErr != nil {
		t.Error("Can't read buffer due to error:", bufferReadErr)

		return
	}

	if bufferReadLen != len(readBuf) {
		t.Errorf("Can't read all data. Expected %d, readed %d",
			len(readBuf), bufferReadLen)

		return
	}

	if string(readBuf) != string(testData[4097:8194]) {
		t.Errorf("Failed to read correct data. Expecting %d, got %d",
			testData[4097:8194], readBuf)

		return
	}

	if source.rCount != 3 {
		t.Errorf("We should read from source 3 times, got %d",
			source.rCount)

		return
	}

	// Last read:
	// We still trying to read 4097 bytes data out, however,
	// there is only 12288 bytes total to read and we already
	// readed 8194 bytes, so there will be only 4094 left.
	// Over read will cause an EOF error
	bufferReadLen, bufferReadErr = bufferBuffer.Read(readBuf)

	if bufferReadErr != io.EOF {
		t.Errorf("Expected over read error to be EOF, got \"%s\" instead",
			bufferReadErr)

		return
	}
}

func TestBufferReadByteByByte(t *testing.T) {
	testData := testGetRandomBytes(testBufferSize * 3)
	source := &dummyByteBuffer{
		r:      bytes.NewBuffer(nil),
		rCount: 0,
		wCount: 0,
	}

	source.Write(testData)

	bufferBuffer := NewBuffer(testBufferSize, source)
	readBuf := make([]byte, 1)
	readIndex := 0
	remainSize := testBufferSize

	for {
		rLen, rErr := bufferBuffer.Read(readBuf)

		if rErr == nil {
			remainSize--

			if bufferBuffer.Len() != remainSize {
				t.Errorf("Expecting buffer length to be %d, got %d",
					remainSize, bufferBuffer.Len())

				return
			}

			if remainSize <= 0 {
				remainSize = testBufferSize
			}

			if rLen != 1 {
				t.Errorf("Expecting read length to be 1, got %d", rLen)

				return
			}

			if testData[readIndex] != readBuf[0] {
				t.Errorf("Failed to read expected bytes. Expecting %d, got %d",
					testData[readIndex], readBuf[0])

				return
			}

			readIndex++

			continue
		}

		if rErr == io.EOF {
			break
		}

		t.Error("Unexpected error:", rErr)
	}

	if source.rCount != 4 {
		t.Errorf("Source should only be readed 4 times, got %d instead",
			source.rCount)

		return
	}
}

func TestBufferReadSmallBufferByteByByte(t *testing.T) {
	testData := testGetRandomBytes(testBufferSize * 3)
	source := &dummyByteBuffer{
		r:      bytes.NewBuffer(nil),
		rCount: 0,
		wCount: 0,
	}

	source.Write(testData)

	bufferBuffer := NewBuffer(1, source)
	readBuf := make([]byte, 1)
	readIndex := 0

	for {
		rLen, rErr := bufferBuffer.Read(readBuf)

		if rErr == nil {
			if bufferBuffer.Len() != 0 {
				t.Errorf("Expecting buffer length to be %d, got %d",
					0, bufferBuffer.Len())

				return
			}

			if rLen != 1 {
				t.Errorf("Expecting read length to be 1, got %d", rLen)

				return
			}

			if testData[readIndex] != readBuf[0] {
				t.Errorf("Failed to read expected bytes. Expecting %d, got %d",
					testData[readIndex], readBuf[0])

				return
			}

			readIndex++

			continue
		}

		if rErr == io.EOF {
			break
		}

		t.Error("Unexpected error:", rErr)
	}

	if source.rCount != (testBufferSize*3)+1 {
		t.Errorf("Source should only be readed %d times, got %d instead",
			(testBufferSize*3)+1, source.rCount)

		return
	}
}

func BenchmarkBufferReadBaseline(b *testing.B) {
	testData := testGetRandomBytes(b.N)
	source := bytes.NewBuffer(nil)

	source.Write(testData)

	readBuf := make([]byte, 1)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rLen, rErr := source.Read(readBuf)

		if rErr != nil {
			b.Error("Unexpected error:", rErr)

			return
		}

		if rLen != 1 {
			b.Error("Unexpected length:", rLen)

			return
		}
	}

	_, srcReadErr := source.Read(readBuf)

	if srcReadErr != io.EOF {
		b.Error("Data is not completely readed from source:")

		return
	}
}

func BenchmarkBufferRead(b *testing.B) {
	testData := testGetRandomBytes(b.N)
	source := bytes.NewBuffer(nil)

	source.Write(testData)

	bufferBuffer := NewBuffer(testBufferSize, source)
	readBuf := make([]byte, 1)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rLen, rErr := bufferBuffer.Read(readBuf)

		if rErr != nil {
			b.Error("Unexpected error:", rErr)

			return
		}

		if rLen != 1 {
			b.Error("Unexpected length:", rLen)

			return
		}
	}

	_, srcReadErr := source.Read(readBuf)

	if srcReadErr != io.EOF {
		b.Error("Data is not completely readed from source:")

		return
	}
}
