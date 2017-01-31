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

type destinationPole struct {
	Head *destination
	Tail *destination
}

type destination struct {
	name       string
	pole       *destinationPole
	transports *transports
	next       *destination
	prev       *destination
}

func (d *destination) Attach() {
	if d.pole.Head == nil {
		d.pole.Head = d
		d.pole.Tail = d

		return
	}

	d.pole.Head.prev = d

	d.prev = nil
	d.next = d.pole.Head

	d.pole.Head = d
}

func (d *destination) Delete() {
	if d == d.pole.Head {
		d.pole.Head = d.next
	}

	if d == d.pole.Tail {
		d.pole.Tail = d.prev
	}

	if d.prev != nil {
		d.prev.next = d.next
	}

	if d.next != nil {
		d.next.prev = d.prev
	}

	d.prev = nil
	d.next = nil
}

func (d *destination) Bump() {
	if d == d.pole.Head {
		return
	}

	d.Delete()

	d.pole.Head.prev = d
	d.next = d.pole.Head

	d.prev = nil

	d.pole.Head = d
}
