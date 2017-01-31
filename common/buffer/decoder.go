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

import "io"

// Decoder read data from the reader and decodes it before output
type Decoder interface {
	Read(p []byte) (int, error)
}

// decoder implements Decoder
type decoder struct {
	buffer       []byte
	shadowBuffer []byte
	reader       io.Reader
	decoder      func(input []byte) ([]byte, error)
	eof          error
}

// NewDecoder creates a new Decoder
func NewDecoder(
	size int,
	reader io.Reader,
	d func(input []byte) ([]byte, error),
) Decoder {
	return &decoder{
		reader:       reader,
		buffer:       make([]byte, 0, size),
		shadowBuffer: make([]byte, size, size),
		decoder:      d,
		eof:          nil,
	}
}

// Read read decoded data
func (b *decoder) Read(p []byte) (int, error) {
	lastRemain := len(b.buffer)
	neededLen := len(p)

	if lastRemain >= neededLen {
		copy(p, b.buffer[:neededLen])
		b.buffer = b.buffer[neededLen:]

		return neededLen, nil
	}

	if b.eof != nil {
		if lastRemain == 0 {
			return 0, b.eof
		}

		copy(p[:lastRemain], b.buffer[:lastRemain])
		b.buffer = b.buffer[:0]

		return lastRemain, nil
	}

	neededShadowBufSize := neededLen - lastRemain

	if neededShadowBufSize > len(b.shadowBuffer) {
		b.shadowBuffer = make(
			[]byte, neededShadowBufSize, neededShadowBufSize)
	}

	sLen, sErr := b.reader.Read(b.shadowBuffer[:neededShadowBufSize])

	if sErr == nil {
		decoded, decodeErr := b.decoder(b.shadowBuffer[:sLen])

		if decodeErr != nil {
			return 0, decodeErr
		}

		b.buffer = append(b.buffer, decoded...)

		lastRemain += len(decoded)

		if lastRemain > neededLen {
			copy(p, b.buffer[:neededLen])
			b.buffer = b.buffer[neededLen:]

			return neededLen, nil
		}

		copy(p, b.buffer[:lastRemain])
		b.buffer = b.buffer[:0]

		return lastRemain, nil
	}

	if sErr == io.EOF && lastRemain > 0 {
		b.eof = sErr

		return b.Read(p)
	}

	return 0, sErr
}
