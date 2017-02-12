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

package transporter

import "io"

// wrapped will wraps given common.ReadWriteCloser, and force them
// to return Transporter error
type wrapped struct {
	io.ReadWriteCloser
}

// Read reads data from ReadWriteCloser and convert returned error
// to Transporter error
func (w *wrapped) Read(b []byte) (int, error) {
	rLen, rErr := w.ReadWriteCloser.Read(b)

	if rErr == nil {
		return rLen, nil
	}

	return rLen, WrapError(rErr)
}

// Write write data to ReadWriteCloser and convert returned error
// to Transporter error
func (w *wrapped) Write(b []byte) (int, error) {
	rLen, rErr := w.ReadWriteCloser.Write(b)

	if rErr == nil {
		return rLen, nil
	}

	return rLen, WrapError(rErr)
}

// Close closes ReadWriteCloser and convert returned error to
// Transporter error
func (w *wrapped) Close() error {
	rErr := w.ReadWriteCloser.Close()

	if rErr == nil {
		return nil
	}

	return WrapError(rErr)
}
