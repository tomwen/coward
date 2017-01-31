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

// Pair is a pair of buffer (Buffer + ExtendedBuffer)
type Pair struct {
	Buffer         Buffer
	ExtendedBuffer Extended
}

// PairSlice is slice of the buffers in the BufferPair
type PairSlice struct {
	Buffer         []byte
	ExtendedBuffer []byte
}

// Slice returns the slice of the buffers in the BufferPair
func (b *Pair) Slice() PairSlice {
	return PairSlice{
		Buffer:         b.Buffer[:],
		ExtendedBuffer: b.ExtendedBuffer[:],
	}
}
