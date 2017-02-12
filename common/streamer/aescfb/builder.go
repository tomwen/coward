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

package aescfb

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"io"

	"github.com/nickrio/coward/common/codec"
	"github.com/nickrio/coward/common/types"
)

// builder will build an AES-CFB crypter
type builder struct {
	key         []byte
	readRand    types.Random
	writeRand   types.Random
	blockCipher cipher.Block
	head        [HeadSize]byte
}

// New creates a new AES-CFB chiper
func New(key []byte) (codec.Streamer, error) {
	readRandom, readRandomErr := types.NewRandom(32)

	if readRandomErr != nil {
		return nil, readRandomErr
	}

	writeRandom, writeRandomErr := types.NewRandom(32)

	if writeRandomErr != nil {
		return nil, writeRandomErr
	}

	blockCipher, blockCipherErr := aes.NewCipher(key)

	if blockCipherErr != nil {
		return nil, blockCipherErr
	}

	return &builder{
		key:         key,
		readRand:    readRandom,
		writeRand:   writeRandom,
		blockCipher: blockCipher,
		head:        [HeadSize]byte{},
	}, nil
}

// OverheadSize returns the  maximum size of overhead
// if this driver
func (c *builder) OverheadSize() int {
	return PaddingMaxSize + MetaSize + SizeSize + PaddingMaxSize
}

// New creates an empty crypter
func (c *builder) New() (codec.Stream, error) {
	rLen, rErr := rand.Read(c.head[:])

	if rErr != nil {
		return nil, rErr
	}

	if rLen != HeadSize {
		return nil, ErrFailedFullyReadHead
	}

	return &crypter{
		head:    c.head,
		encoder: cipher.NewCFBEncrypter(c.blockCipher, c.head[:rLen]),
		padding: padding{
			random:     c.writeRand,
			cachedByte: [2]byte{},
			max:        16,
		},
		readMac:     hmac.New(sha256.New, c.key),
		writeMac:    hmac.New(sha256.New, c.key),
		data:        [codec.BufferSize]byte{},
		sizeEnBytes: [SizeSize]byte{},
		sizeDeBytes: [SizeSize]byte{},
		metaBytes:   [MetaSize]byte{},
	}, nil
}

// Init initials a crypter from reader data
func (c *builder) Init(reader io.Reader) (codec.Stream, error) {
	rLen, rErr := reader.Read(c.head[:])

	if rErr != nil {
		return nil, rErr
	}

	if rLen != HeadSize {
		return nil, ErrFailedFullyWriteHead
	}

	return &crypter{
		head:    c.head,
		decoder: cipher.NewCFBDecrypter(c.blockCipher, c.head[:rLen]),
		padding: padding{
			random:     c.readRand,
			cachedByte: [2]byte{},
			max:        16,
		},
		readMac:     hmac.New(sha256.New, c.key),
		writeMac:    hmac.New(sha256.New, c.key),
		data:        [codec.BufferSize]byte{},
		sizeEnBytes: [SizeSize]byte{},
		sizeDeBytes: [SizeSize]byte{},
		metaBytes:   [MetaSize]byte{},
	}, nil
}
