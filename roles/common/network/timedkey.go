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

package network

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"time"
)

var (
	// ErrKeyResultTooShort is throwed when key result is too short
	ErrKeyResultTooShort = errors.New(
		"Key result is too short")

	// ErrKeyFailedToGenerateTimeBytes is throwed when we failed to
	// generate time bytes
	ErrKeyFailedToGenerateTimeBytes = errors.New(
		"Fail to generate time bytes for current key")
)

var (
	emptyKeytime = [8]byte{}
)

const (
	// KeyExpireDuration Key will be valid with in this time
	KeyExpireDuration = 5 * time.Second
)

// TimedKey is the password used for encryption and decryption
type TimedKey []byte

// Get current key
func (k TimedKey) Get(getTime time.Time, resultLen int) ([]byte, error) {
	nowByte := [8]byte{}

	nowInt := uint64(getTime.Truncate(KeyExpireDuration).Unix())

	binary.BigEndian.PutUint64(nowByte[:], nowInt)

	if nowByte == emptyKeytime {
		return nil, ErrKeyFailedToGenerateTimeBytes
	}

	mac := hmac.New(sha256.New, k)

	_, wErr := mac.Write(nowByte[:])

	if wErr != nil {
		return nil, wErr
	}

	result := mac.Sum(nil)

	if len(result) < resultLen {
		return nil, ErrKeyResultTooShort
	}

	return result[:resultLen], nil
}
