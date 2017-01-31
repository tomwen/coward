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

import "github.com/nickrio/coward/roles/common/buffer"

// Buffer is a pair buffer for Client and Server
type Buffer struct {
	Client buffer.Pair
	Server buffer.Pair
}

// Slice is two Buffer PairSlices for Client and Server
type Slice struct {
	Client buffer.PairSlice
	Server buffer.PairSlice
}

// Slice build two Buffer PairSlices for Client and Server
func (b *Buffer) Slice() Slice {
	return Slice{
		Client: b.Client.Slice(),
		Server: b.Server.Slice(),
	}
}
