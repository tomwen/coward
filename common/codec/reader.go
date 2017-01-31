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
	"io"

	"github.com/nickrio/coward/common/buffer"
)

// reader read streamerped-text into plain-text data
type reader struct {
	common

	streamer  Streamer
	stream    Stream
	buffer    buffer.Buffer
	decodeBuf buffer.Decoder
	outputBuf buffer.Generater
}

// NewReader creates a new Reader
func NewReader(streamer Streamer, r io.Reader) io.Reader {
	return &reader{
		streamer: streamer,
		buffer:   buffer.NewBuffer(BufferSize, r),
	}
}

// Read and unstreamer data from the reader source
func (r *reader) Read(p []byte) (int, error) {
	if !r.ready {
		stream, streamErr := r.streamer.Init(r.buffer)

		if streamErr != nil {
			return 0, streamErr
		}

		r.ready = true
		r.stream = stream
		r.decodeBuf = buffer.NewDecoder(
			BufferSize,
			r.buffer,
			func(input []byte) ([]byte, error) {
				return r.stream.Decode(input)
			})

		r.outputBuf = buffer.NewGenerater(BufferSize, func() ([]byte, error) {
			return r.stream.Unstream(r.decodeBuf)
		})
	}

	return r.outputBuf.Read(p)
}
