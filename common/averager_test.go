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

package common

import "testing"

func testAverager(t *testing.T, avg Averager) {
	if avg.Get() != 0.0 {
		t.Error("Empty average must return 0 value")

		return
	}

	avg.Add(10.0)

	if avg.Get() != 10.0 {
		t.Error("Averager: 10 / 1 != 10")

		return
	}

	avg.Add(10.0)

	if avg.Get() != 10.0 {
		t.Error("Averager: 10 + 10 / 2 != 10")

		return
	}

	// At this step, the first 10 will be cleared from average record,
	// leaving the second 10 and new added 0, so the average will be 5
	avg.Add(0.0)

	if avg.Get() != 5.0 {
		t.Error("Averager: 10 + 0 / 2 != 5")

		return
	}

	// Add 10 and make it weight a whole averager
	avg.AddWithWeight(10, func(average float64, size int) int {
		return size
	})

	if avg.Get() != 10.0 {
		t.Error("Averager: add 10 with whole averager " +
			"weight not resulting a 10 averager")

		return
	}
}

func TestAverager(t *testing.T) {
	testAverager(t, NewAverager(2))
}

func TestLockedAverager(t *testing.T) {
	testAverager(t, NewLockedAverager(2))
}
