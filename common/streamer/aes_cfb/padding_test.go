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

package aesCFB

import (
	"bytes"
	"testing"

	"github.com/nickrio/coward/common/types"
)

func TestPadding(t *testing.T) {
	data := make([]byte, 0, 1024)
	buf := bytes.NewBuffer(data)
	maxTestLimit := 256

	rand, randErr := types.NewRandom(32)

	if randErr != nil {
		t.Error("Failed to create random generater for the test:", randErr)

		return
	}

	p := padding{
		random:     rand,
		cachedByte: [2]byte{},
		max:        16,
	}

	for {
		wErr := p.WriteIn(buf)

		if wErr != nil {
			t.Error("Failed to write padding due to error:", wErr)

			return
		}

		if buf.Len() == 0 {
			t.Error("Failed to write padding")

			return
		}

		walkErr := p.Walkthrough(buf)

		if walkErr != nil {
			t.Error("Failed to read padding due to error:", walkErr)

			return
		}

		if buf.Len() != 0 {
			t.Error("Failed to read padding")

			return
		}

		maxTestLimit--

		if maxTestLimit <= 0 {
			break
		}
	}
}
