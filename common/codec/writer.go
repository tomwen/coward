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
	"io"
	"math"
)

// Writer write plain-text into wrapped-text data
type writer struct {
	common

	streamer Streamer
	stream   Stream
	writer   io.Writer
	buffer   *bytes.Buffer
}

// NewWriter creates a new Writer
func NewWriter(streamer Streamer, w io.Writer) io.Writer {
	return &writer{
		streamer: streamer,
		writer:   w,
	}
}

// Write writes wrapped data to the writer
func (w *writer) Write(p []byte) (int, error) {
	if !w.ready {
		if BufferSize <= w.streamer.OverheadSize() {
			return 0, ErrStreamOverheadTooLarge
		}

		stream, streamErr := w.streamer.New()

		if streamErr != nil {
			return 0, streamErr
		}

		w.stream = stream

		headBuildErr := w.stream.Head(w.writer)

		if headBuildErr != nil {
			return 0, headBuildErr
		}

		w.ready = true
		w.buffer = bytes.NewBuffer(
			make([]byte, 0, BufferSize+w.streamer.OverheadSize()))
	}

	dataLength := len(p)
	dataParagraphSize := BufferSize - w.streamer.OverheadSize()
	paragraphs := math.Ceil(float64(dataLength) / float64(dataParagraphSize))
	paragraphStart := 0

	for part := float64(0); part < paragraphs; part++ {
		w.buffer.Reset()

		paragraphEnd := paragraphStart + dataParagraphSize

		if paragraphEnd > dataLength {
			paragraphEnd = dataLength
		}

		wrapErr := w.stream.Stream(
			p[paragraphStart:paragraphEnd], w.buffer)

		if wrapErr != nil {
			return 0, wrapErr
		}

		encodeErr := w.stream.Encode(w.buffer.Bytes(), w.writer)

		if encodeErr != nil {
			return 0, encodeErr
		}

		paragraphStart = paragraphEnd
	}

	return len(p), nil
}
