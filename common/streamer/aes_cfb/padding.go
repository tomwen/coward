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
	"errors"
	"io"
	"math"

	"github.com/nickrio/coward/common/types"
)

var (
	// ErrPaddingFailedToWritePaddingLength is throwed when we can't
	// write the padding data
	ErrPaddingFailedToWritePaddingLength = errors.New(
		"Failed to write padding")

	// ErrPaddingFailedToReadPaddingLength is throwed when we can't
	// read the padding data
	ErrPaddingFailedToReadPaddingLength = errors.New(
		"Failed to read padding length")

	// ErrPaddingFailedWalkout is throwed when we can't found the
	// ending of the padding
	ErrPaddingFailedWalkout = errors.New(
		"Failed to read padding length")
)

type padding struct {
	random     types.Random
	cachedByte [2]byte
	max        byte
}

func (p *padding) WriteIn(w io.Writer) error {
	rMaxNum, rMaxErr := p.random.GetMax(p.max)

	if rMaxErr != nil {
		return rMaxErr
	}

	if rMaxNum < 1 {
		rMaxNum++
	}

	rMaxNum--

	for randIdx := byte(0); randIdx < rMaxNum; randIdx++ {
		rNum, rErr := p.random.Get()

		if rErr != nil {
			return rErr
		}

		p.cachedByte[0] = rNum | 1

		wLen, wErr := w.Write(p.cachedByte[:1])

		if wErr != nil {
			return wErr
		}

		if wLen != 1 {
			return ErrPaddingFailedToWritePaddingLength
		}
	}

	rNum, rErr := p.random.Get()

	if rErr != nil {
		return rErr
	}

	p.cachedByte[0] = rNum &^ 1

	wLen, wErr := w.Write(p.cachedByte[:1])

	if wErr != nil {
		return wErr
	}

	if wLen != 1 {
		return ErrPaddingFailedToWritePaddingLength
	}

	return nil
}

func (p *padding) Walkthrough(r io.Reader) error {
	totalReaded := 0

	for {
		rLen, rErr := r.Read(p.cachedByte[1:])

		if rErr != nil {
			return rErr
		}

		if rLen != 1 {
			return ErrPaddingFailedToReadPaddingLength
		}

		if totalReaded > int(p.max) {
			return ErrPaddingFailedWalkout
		}

		if totalReaded >= int(math.MaxUint8) {
			return ErrPaddingFailedWalkout
		}

		totalReaded += rLen

		if p.cachedByte[1]&1 == 0 {
			break
		}
	}

	return nil
}
