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
	"crypto/rand"
	"errors"
)

var (
	// ErrUnexpectedRandomByteLength when outputing byte
	// is unexpected in length
	ErrUnexpectedRandomByteLength = errors.New(
		"Unexpected Random Byte Length")

	// ErrRandomByteLengthMustGreaterThanZero must be greater
	// than zero
	ErrRandomByteLengthMustGreaterThanZero = errors.New(
		"Random byte length must greater than zero")
)

// Random generate random numbers
type Random struct {
	bytes []byte
	index int
	max   int
}

// NewRandom Get a new Random number generater
func NewRandom(bufferLength int) (Random, error) {
	r := Random{
		bytes: make([]byte, bufferLength),
		index: 0,
		max:   bufferLength,
	}

	if bufferLength < 1 {
		return Random{}, ErrRandomByteLengthMustGreaterThanZero
	}

	gErr := r.generate()

	if gErr != nil {
		return Random{}, gErr
	}

	return r, nil
}

func (r *Random) generate() error {
	rLen, rErr := rand.Read(r.bytes)

	if rErr != nil {
		return rErr
	}

	if rLen <= 0 {
		return ErrUnexpectedRandomByteLength
	}

	if rLen != r.max {
		return ErrUnexpectedRandomByteLength
	}

	return nil
}

// Get get a rand byte from the generater
func (r *Random) Get() (byte, error) {
	if r.index >= r.max {
		gErr := r.generate()

		if gErr != nil {
			return 0, gErr
		}

		r.index = 0
	}

	result := r.bytes[r.index]

	r.index++

	return result, nil
}

// GetMax get a random number that can't not greater than or equals to max
func (r *Random) GetMax(max byte) (byte, error) {
	rNum, rErr := r.Get()

	if rErr != nil {
		return 0, rErr
	}

	return rNum % max, nil
}
