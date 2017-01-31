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

// Buffer is a byte buffer stored in a buffer slice, always trying to fill
// itself after returning buffer data
type Buffer interface {
	io.Reader
	Len() int
}

// Buffer is a byte buffer that assign and retrieve in circle
type buffer struct {
	size    int
	src     io.Reader
	start   int
	end     int
	flipped bool
	buf     []byte
}

// NewBuffer creates a buffer buffer
func NewBuffer(size int, source io.Reader) Buffer {
	return &buffer{
		size:    size,
		src:     source,
		start:   0,
		end:     0,
		flipped: false,
		buf:     make([]byte, size),
	}
}

func (r *buffer) fillBufferFromSource() (int, error) {
	var rLen int
	var rErr error

	// flipped means the buffer will be something like:
	// [D][D][E][ ][ ][ ][ ][S][D][D]
	// |-HEAD-|-   SPACE   -|-TAIL-|
	if r.flipped {
		rLen, rErr = r.src.Read(r.buf[r.end:r.start])

		if rErr == nil {
			r.end += rLen

			return rLen, nil
		}

		return rLen, rErr
	}

	// Un-flipped will be like:
	// [ ][ ][S][D][D][D][D][E][ ][ ]
	// |-HEAD-|-   FILLED  -|-TAIL-|
	// A little rules here is, we only read to TAIL if there is any
	// freespace available. Because of the risk of IO waiting block
	if r.size-r.end > 0 {
		rLen, rErr = r.src.Read(r.buf[r.end:r.size])

		if rErr == nil {
			r.end += rLen

			return rLen, nil
		}

		return rLen, rErr
	}

	// If tail is fully filled however, then we must filling the HEAD
	// and mark the buffer as flipped
	rLen, rErr = r.src.Read(r.buf[0:r.start])

	if rErr == nil {
		r.end = rLen
		r.flipped = true

		return rLen, nil
	}

	return rLen, rErr
}

// Retrive data from buffer and update length information
// Notice here we assume the readLen parameter will never go beyond actual
// buffer length
func (r *buffer) retrieveFromBuffer(b []byte, readLen int) int {
	retrieveLen := 0

	if !r.flipped {
		// Un-flipped means the buffer will be like:
		// [0][S][D][D][D][D][D][E][0][0]
		//    |-      DATA      -|
		retrieveLen = r.end - r.start

		if retrieveLen <= 0 {
			return 0
		}

		if retrieveLen > readLen {
			retrieveLen = readLen
		}

		copy(b[0:retrieveLen], r.buf[r.start:r.start+retrieveLen])

		r.start += retrieveLen

		return retrieveLen
	}

	// flipped means the buffer will be something like:
	// [D][D][E][ ][ ][ ][ ][S][D][D]
	// |-HEAD-|             |-TAIL-|
	tailLen := r.size - r.start

	// See if we have to retrive both two segments (Head and tail)
	if tailLen >= readLen {
		retrieveLen = readLen

		copy(b[0:retrieveLen], r.buf[r.start:r.start+retrieveLen])

		r.start += retrieveLen

		if r.start == r.size {
			r.start = 0
			r.flipped = false
		}

		return retrieveLen
	}

	retrieveLen = tailLen

	copy(b[0:retrieveLen], r.buf[r.start:r.size])

	r.start = 0
	r.flipped = false

	remainReadLen := readLen - retrieveLen
	headCopyLen := r.end - tailLen

	if headCopyLen > remainReadLen {
		headCopyLen = remainReadLen
	}

	copy(b[retrieveLen:retrieveLen+headCopyLen],
		r.buf[0:headCopyLen])

	r.start += headCopyLen
	retrieveLen += headCopyLen

	return retrieveLen
}

func (r *buffer) Read(b []byte) (int, error) {
	bLen := len(b)
	rLen := r.Len()

	// If we already has data to fill in to b,
	// no reason to read the source
	if rLen >= bLen {
		return r.retrieveFromBuffer(b, bLen), nil
	}

	// If the data is not enough to fill b
	// fill the first part of it first
	filledLen := r.retrieveFromBuffer(b[0:], bLen)

	for {
		if filledLen >= bLen {
			return filledLen, nil
		}

		_, sourceReadErr := r.fillBufferFromSource()

		if sourceReadErr != nil {
			return 0, sourceReadErr
		}

		filledLen += r.retrieveFromBuffer(b[filledLen:], len(b[filledLen:]))
	}
}

func (r *buffer) Len() int {
	if r.flipped {
		if r.start == r.end {
			return r.size
		}

		return r.end + (r.size - r.start)
	}

	return r.end - r.start
}
