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

import (
	"bytes"
	"crypto/rand"
	"errors"
	"io"
	"testing"
)

type testDummyForReaderWrapper struct {
	textData  []byte
	metaData  []byte
	frontData []byte
	headData  []byte
	size      uint16
	font      []byte
	text1     []byte
	text2     []byte
	sizeb     []byte
	sizeData  []byte
	meta      []byte
	readed    []byte
}

func (t *testDummyForReaderWrapper) Head(writer io.Writer) error {
	_, wErr := writer.Write(t.headData)

	return wErr
}

func (t *testDummyForReaderWrapper) Stream(data []byte, w io.Writer) error {
	w.Write(t.frontData)
	w.Write(t.textData)
	w.Write(t.sizeData)
	w.Write(t.metaData)
	w.Write(data)
	w.Write(t.textData)

	return nil
}

func (t *testDummyForReaderWrapper) Encode(input []byte, w io.Writer) error {
	for idx := range input {
		input[idx]++
	}

	inputLen, inputErr := w.Write(input)

	if inputErr != nil {
		return inputErr
	}

	if inputLen != len(input) {
		return errors.New("Can't write all encoded contents to writer")
	}

	return nil
}

func (t *testDummyForReaderWrapper) Decode(
	input []byte) (decoded []byte, decodeErr error) {
	for idx := range input {
		input[idx]--
	}

	return input, nil
}

func (t *testDummyForReaderWrapper) Unstream(r io.Reader) ([]byte, error) {
	// Font
	fontReadLen, fontReadErr := r.Read(t.font)

	if fontReadErr != nil {
		return []byte{}, fontReadErr
	}

	if string(t.font[:fontReadLen]) != string(t.frontData) {
		return []byte{}, errors.New("Invalid front data")
	}

	// Text 1
	text1ReadLen, text1ReadErr := r.Read(t.text1)

	if text1ReadErr != nil {
		return []byte{}, text1ReadErr
	}

	if string(t.text1[:text1ReadLen]) != string(t.textData) {
		return []byte{}, errors.New("Invalid text data 1")
	}

	// Size
	sizeReadLen, sizeReadErr := r.Read(t.sizeb)

	if sizeReadErr != nil {
		return []byte{}, sizeReadErr
	}

	if string(t.sizeb[:sizeReadLen]) != string(t.sizeData) {
		return []byte{}, errors.New("Invalid size data")
	}

	// Meta
	metaReadLen, metaReadErr := r.Read(t.meta)

	if metaReadErr != nil {
		return []byte{}, metaReadErr
	}

	if string(t.meta[:metaReadLen]) != string(t.metaData) {
		return []byte{}, errors.New("Invalid meta data")
	}

	// Data
	dataReadLen, dataReadErr := r.Read(t.readed)

	if dataReadErr != nil {
		return []byte{}, dataReadErr
	}

	// Text 2
	text2ReadLen, text2ReadErr := r.Read(t.text2)

	if text2ReadErr != nil {
		return []byte{}, text2ReadErr
	}

	if string(t.text1[:text2ReadLen]) != string(t.textData) {
		return []byte{}, errors.New("Invalid text data 2")
	}

	return t.readed[:dataReadLen], nil
}

type testDummyForReaderWrap struct {
	size uint16
}

func (t *testDummyForReaderWrap) OverheadSize() int {
	return 16 + 4 + 10 + 16
}

func (t *testDummyForReaderWrap) New() (Stream, error) {
	return &testDummyForReaderWrapper{
		textData:  []byte("PADDINGPADDINGPP"),
		metaData:  []byte("METAMETAMM"),
		frontData: []byte("WRAPPING"),
		headData:  []byte("HEADHEADHEADHEAD"),
		sizeData:  []byte("    "),
		size:      t.size,
		font:      make([]byte, 8),
		text1:     make([]byte, 16),
		text2:     make([]byte, 16),
		sizeb:     make([]byte, 4),
		meta:      make([]byte, 10),
		readed:    make([]byte, t.size),
	}, nil
}

func (t *testDummyForReaderWrap) Init(reader io.Reader) (Stream, error) {
	headData := make([]byte, 16)

	rLen, rErr := reader.Read(headData)

	if rErr != nil {
		return nil, rErr
	}

	if rLen != len(headData) {
		return nil, errors.New("Invalid head: Unexpected size")
	}

	for idx := range headData[:rLen] {
		headData[idx]--
	}

	return &testDummyForReaderWrapper{
		textData:  []byte("PADDINGPADDINGPP"),
		metaData:  []byte("METAMETAMM"),
		frontData: []byte("WRAPPING"),
		sizeData:  []byte("    "),
		headData:  headData[:rLen],
		size:      t.size,
		font:      make([]byte, 8),
		text1:     make([]byte, 16),
		text2:     make([]byte, 16),
		sizeb:     make([]byte, 4),
		meta:      make([]byte, 10),
		readed:    make([]byte, t.size),
	}, nil
}

func TestReader(t *testing.T) {
	data := []byte(
		"HEADHEADHEADHEADWRAPPING" +
			"PADDINGPADDINGPP    METAMETAMMDATAPADDINGPADDINGPP",
	)

	for idx := range data {
		data[idx]++
	}

	buf := bytes.NewBuffer(data)

	cipher := &testDummyForReaderWrap{
		size: 4,
	}

	reader := NewReader(cipher, buf)

	readed := make([]byte, 256)

	rLen, rErr := reader.Read(readed)

	if rErr != nil {
		t.Error("Can't read due to error:", rErr)

		return
	}

	if string(readed[:rLen]) != "DATA" {
		t.Error("Can't read correct data out of read source")

		return
	}

	rLen, rErr = reader.Read(readed)

	if rErr != io.EOF {
		t.Error("Failed to read EOF")

		return
	}

	if rLen != 0 {
		t.Error("EOF length must be 0")

		return
	}
}

func TestWriteRead(t *testing.T) {
	data := make([]byte, 0, 1<<16)
	expecting := make([]byte, 0, 1<<16)
	buf := bytes.NewBuffer(data)

	loopTimes := 1 << 16 / 1024
	randomData := make([]byte, 1024)

	wrap := &testDummyForReaderWrap{
		size: 1024,
	}

	writer := NewWriter(wrap, buf)
	reader := NewReader(wrap, buf)

	for i := 0; i < loopTimes; i++ {
		_, randReadErr := rand.Read(randomData)

		if randReadErr != nil {
			t.Error("Can't generate random testing data")

			return
		}

		writer.Write(randomData)

		expecting = append(expecting, randomData...)
	}

	expectingIdxOffset := 0

	for {
		readLen, readErr := reader.Read(randomData)

		if readErr != nil {
			if readErr == io.EOF {
				break
			}

			t.Error("Can't read due to error:", readErr)

			return
		}

		if string(expecting[expectingIdxOffset:expectingIdxOffset+readLen]) !=
			string(randomData[:readLen]) {
			t.Errorf("Can't read expected data."+
				"\r\nExpecting %v (%v %v %v), \r\n      Got %v (%v)",
				expecting[expectingIdxOffset:expectingIdxOffset+readLen],
				len(expecting[expectingIdxOffset:expectingIdxOffset+readLen]),
				expectingIdxOffset, expectingIdxOffset+readLen,
				randomData[:readLen],
				len(randomData[:readLen]))

			return
		}

		expectingIdxOffset += readLen
	}
}

func BenchmarkWriteReader(b *testing.B) {
	data := make([]byte, 0, 4096)
	buf := bytes.NewBuffer(data)
	wrap := &testDummyForReaderWrap{
		size: 4050,
	}

	writer := NewWriter(wrap, buf)
	reader := NewReader(wrap, buf)

	writeBuf := make([]byte, 4096)
	readBuf := make([]byte, 4096)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, wErr := writer.Write(writeBuf)

		if wErr != nil {
			b.Error("Can't write due to error:", wErr)

			return
		}

		for {
			_, rErr := reader.Read(readBuf)

			if rErr == nil {
				continue
			}

			if rErr == io.EOF {
				b.StopTimer()

				buf.Reset()

				b.StartTimer()

				break
			}

			b.Error("Can't read due to error:", rErr)

			return
		}
	}
}
