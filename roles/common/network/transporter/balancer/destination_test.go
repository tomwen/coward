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

package balancer

import (
	"fmt"
	"strings"
	"testing"
)

func testDestinationTraverse(pole *destinationPole) []string {
	resultOrder := []string{}

	current := pole.Head

	for {
		if current == nil {
			break
		}

		resultOrder = append(resultOrder, current.name)

		current = current.next
	}

	current = pole.Tail

	for {
		if current == nil {
			break
		}

		resultOrder = append(resultOrder, current.name)

		current = current.prev
	}

	return resultOrder
}

func TestDestinationAttach(t *testing.T) {
	pole := destinationPole{
		Head: nil,
		Tail: nil,
	}

	for i := 0; i < 10; i++ {
		dest := &destination{
			name:       fmt.Sprintf("D%d", i),
			pole:       &pole,
			transports: nil,
			next:       nil,
			prev:       nil,
		}

		dest.Attach()
	}

	if strings.Join(testDestinationTraverse(&pole), " ") !=
		"D9 D8 D7 D6 D5 D4 D3 D2 D1 D0 D0 D1 D2 D3 D4 D5 D6 D7 D8 D9" {
		t.Errorf("Failed to insert in a right order. Expecting %s, got %s",
			"D9 D8 D7 D6 D5 D4 D3 D2 D1 D0 D0 D1 D2 D3 D4 D5 D6 D7 D8 D9",
			testDestinationTraverse(&pole))

		return
	}
}

func TestDestinationDelete(t *testing.T) {
	var d6 *destination

	pole := destinationPole{
		Head: nil,
		Tail: nil,
	}

	for i := 0; i < 10; i++ {
		dest := &destination{
			name:       fmt.Sprintf("D%d", i),
			pole:       &pole,
			transports: nil,
			next:       nil,
			prev:       nil,
		}

		dest.Attach()

		switch i {
		case 5:
			dest.Delete()

		case 6:
			d6 = dest
		}
	}

	if strings.Join(testDestinationTraverse(&pole), " ") !=
		"D9 D8 D7 D6 D4 D3 D2 D1 D0 D0 D1 D2 D3 D4 D6 D7 D8 D9" {
		t.Errorf("Failed to insert in a right order. Expecting %s, got %s",
			"D9 D8 D7 D6 D4 D3 D2 D1 D0 D0 D1 D2 D3 D4 D6 D7 D8 D9",
			testDestinationTraverse(&pole))

		return
	}

	d6.Delete()

	if strings.Join(testDestinationTraverse(&pole), " ") !=
		"D9 D8 D7 D4 D3 D2 D1 D0 D0 D1 D2 D3 D4 D7 D8 D9" {
		t.Errorf("Failed to insert in a right order. Expecting %s, got %s",
			"D9 D8 D7 D4 D3 D2 D1 D0 D0 D1 D2 D3 D4 D7 D8 D9",
			testDestinationTraverse(&pole))

		return
	}
}

func TestDestinationBump(t *testing.T) {
	var d3 *destination
	var d6 *destination

	pole := destinationPole{
		Head: nil,
		Tail: nil,
	}

	for i := 0; i < 10; i++ {
		dest := &destination{
			name:       fmt.Sprintf("D%d", i),
			pole:       &pole,
			transports: nil,
			next:       nil,
			prev:       nil,
		}

		dest.Attach()

		switch i {
		case 3:
			d3 = dest

		case 6:
			d6 = dest
		}
	}

	d3.Bump()

	if strings.Join(testDestinationTraverse(&pole), " ") !=
		"D3 D9 D8 D7 D6 D5 D4 D2 D1 D0 D0 D1 D2 D4 D5 D6 D7 D8 D9 D3" {
		t.Errorf("Failed to insert in a right order. Expecting %s, got %s",
			"D3 D9 D8 D7 D6 D5 D4 D2 D1 D0 D0 D1 D2 D4 D5 D6 D7 D8 D9 D3",
			testDestinationTraverse(&pole))

		return
	}

	d6.Bump()

	if strings.Join(testDestinationTraverse(&pole), " ") !=
		"D6 D3 D9 D8 D7 D5 D4 D2 D1 D0 D0 D1 D2 D4 D5 D7 D8 D9 D3 D6" {
		t.Errorf("Failed to insert in a right order. Expecting %s, got %s",
			"D6 D3 D9 D8 D7 D5 D4 D2 D1 D0 D0 D1 D2 D4 D5 D7 D8 D9 D3 D6",
			testDestinationTraverse(&pole))

		return
	}

	d6.Delete()

	if strings.Join(testDestinationTraverse(&pole), " ") !=
		"D3 D9 D8 D7 D5 D4 D2 D1 D0 D0 D1 D2 D4 D5 D7 D8 D9 D3" {
		t.Errorf("Failed to insert in a right order. Expecting %s, got %s",
			"D3 D9 D8 D7 D5 D4 D2 D1 D0 D0 D1 D2 D4 D5 D7 D8 D9 D3",
			testDestinationTraverse(&pole))

		return
	}
}
