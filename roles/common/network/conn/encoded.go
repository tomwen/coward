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

package conn

import (
	"io"
	"net"

	"github.com/nickrio/coward/common/codec"
)

// encoded encoded conn data
type encoded struct {
	net.Conn

	reader io.Reader
	writer io.Writer
}

// NewEncoded creates a new Encoded CONN
func NewEncoded(raw net.Conn, cipher codec.Streamer) net.Conn {
	return &encoded{
		Conn:   raw,
		reader: codec.NewReader(cipher, raw),
		writer: codec.NewWriter(cipher, raw),
	}
}

// Read read from conn
func (c *encoded) Read(p []byte) (int, error) {
	return c.reader.Read(p)
}

// Write write to conn
func (c *encoded) Write(p []byte) (int, error) {
	return c.writer.Write(p)
}
