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

package types

import (
	"testing"
)

func TestRandomGet(t *testing.T) {
	randNumberBucket := map[byte]uint{}
	isRepeated := false
	r, rErr := NewRandom(10)

	if rErr != nil {
		t.Error("Error building random number generator:", rErr)

		return
	}

	for i := 0; i < 100; i++ {
		randByte, randErr := r.Get()

		if randErr != nil {
			t.Error("Errored while getting random bytes", randErr)

			break
		}

		if _, found := randNumberBucket[randByte]; found {
			randNumberBucket[randByte]++
		} else {
			randNumberBucket[randByte] = 1
		}
	}

	for _, rCount := range randNumberBucket {
		if rCount < 10 {
			continue
		}

		isRepeated = true

		break
	}

	if isRepeated {
		t.Error("Random number is repeating")

		return
	}
}

func TestRandomGetMax(t *testing.T) {
	r, rErr := NewRandom(10)

	if rErr != nil {
		t.Error("Failed to build random number generator:", rErr)
	}

	for i := 0; i < 1000; i++ {
		n, nErr := r.GetMax(10)

		if nErr != nil {
			t.Error(nErr)

			return
		}

		if n < 10 {
			continue
		}

		t.Error("Result of `GetMax` can't greater than 10")
	}
}

func BenchmarkRandomGet(b *testing.B) {
	r, rErr := NewRandom(10)

	if rErr != nil {
		b.Error(rErr)

		return
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, gotErr := r.Get()

		if gotErr != nil {
			b.Error(gotErr)

			return
		}
	}
}
