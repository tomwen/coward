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

package locked

import "testing"

func TestBool(t *testing.T) {
	b := NewBool(false)
	b2 := NewBool(true)

	if b.Get() {
		t.Error("Expecting the default value of Bool 1 is false, got true")

		return
	}

	if !b2.Get() {
		t.Error("Expecting the default value of Bool 2 is true, got false")

		return
	}

	b.Set(true)
	b2.Set(false)

	if !b.GetSet(false) {
		t.Error("Expecting the updated value of Bool 1 is true, got false")

		return
	}

	if b2.GetSet(true) {
		t.Error("Expecting the updated value of Bool 2 is false, got true")

		return
	}

	if b.Get() {
		t.Error("Expecting the updated value of Bool 1 is false, got true")

		return
	}

	if !b2.Get() {
		t.Error("Expecting the updated value of Bool 2 is true, got false")

		return
	}

	finalB := false
	finalB2 := true

	b.Load(func(current bool) {
		finalB = current
	})

	b2.Load(func(current bool) {
		finalB2 = current
	})

	if finalB {
		t.Error("Expecting the updated value of Bool 1 is false, got true")

		return
	}

	if !finalB2 {
		t.Error("Expecting the updated value of Bool 2 is true, got false")

		return
	}
}
