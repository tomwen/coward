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

// Generater read data from a generater function
type Generater interface {
	Read(p []byte) (int, error)
}

// Generate generates data
type Generate func() ([]byte, error)

// generater implements Generater
type generater struct {
	buffer   []byte
	bufLen   int
	generate Generate
}

// NewGenerater creates a new generater
func NewGenerater(bufSize int, generate Generate) Generater {
	return &generater{
		buffer:   make([]byte, 0, bufSize),
		bufLen:   0,
		generate: generate,
	}
}

// Read reads data from buffer
func (s *generater) Read(p []byte) (int, error) {
	readLen := len(p)

	// If we had can fill the requested data with buffered data, fill it
	if s.bufLen >= readLen {
		copy(p[:readLen], s.buffer[:readLen])
		s.buffer = s.buffer[readLen:]

		s.bufLen -= readLen

		return readLen, nil
	}

	// If not, that means we need to read some data from source
	srcData, srcErr := s.generate()

	if srcErr == nil {
		s.buffer = append(s.buffer, srcData...)
		s.bufLen = len(s.buffer)

		// Retest: Fully fillable?
		if s.bufLen > readLen {
			copy(p[:readLen], s.buffer[:readLen])
			s.buffer = s.buffer[readLen:]

			s.bufLen -= readLen

			return readLen, nil
		}

		oldBufLen := s.bufLen

		copy(p[:s.bufLen], s.buffer[:s.bufLen])

		s.buffer = s.buffer[:0]
		s.bufLen = 0

		return oldBufLen, nil
	}

	if srcErr == io.EOF && s.bufLen > 0 {
		oldBufLen := s.bufLen

		copy(p[:s.bufLen], s.buffer[:s.bufLen])

		s.bufLen = 0

		return oldBufLen, nil
	}

	return 0, srcErr
}
