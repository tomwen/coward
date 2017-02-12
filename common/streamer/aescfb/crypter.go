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
	"crypto/cipher"
	"crypto/hmac"
	"hash"
	"io"

	"github.com/nickrio/coward/common/codec"
	"github.com/nickrio/coward/common/types"
)

// crypter encrypt and decrypt data according
// to the setting
type crypter struct {
	head        [HeadSize]byte
	encoder     cipher.Stream
	decoder     cipher.Stream
	padding     padding
	readMac     hash.Hash
	writeMac    hash.Hash
	data        [codec.BufferSize]byte
	sizeDeBytes [SizeSize]byte
	sizeEnBytes [SizeSize]byte
	metaBytes   [MetaSize]byte
}

// hmac calecute the HMAC using given message and key
func (c *crypter) hmac(mac hash.Hash, message []byte) ([]byte, error) {
	// Stream hmac, keep writing
	wLen, wErr := mac.Write(message)

	if wErr != nil {
		return nil, wErr
	}

	if wLen != len(message) {
		return nil, ErrFailedToWriteHMACToSumer
	}

	hmac := mac.Sum(nil)

	if len(hmac) < MetaSize {
		return nil, ErrFailedHMACTooShort
	}

	return hmac[:MetaSize], nil
}

// hmacEqual check if two HMAC is equal
func (c *crypter) hmacEqual(hmac1 []byte, hmac2 []byte) bool {
	return hmac.Equal(hmac1, hmac2)
}

// Head build the head of this codec driver
func (c *crypter) Head(writer io.Writer) error {
	wLen, wErr := writer.Write(c.head[:])

	if wErr != nil {
		return wErr
	}

	if wLen != HeadSize {
		return ErrFailedFullyWriteHead
	}

	return nil
}

// Stream wrap the data and write result to the writer
func (c *crypter) Stream(input []byte, w io.Writer) error {
	// Font Padding
	fontPaddingErr := c.padding.WriteIn(w)

	if fontPaddingErr != nil {
		return fontPaddingErr
	}

	// Size, notice the input can't larger than 4096, it's the limit set
	// by codec, not here
	encodeableSize := types.EncodableUint16(len(input))

	sizeEncodeErr := encodeableSize.EncodeBytes(c.sizeEnBytes[:])

	if sizeEncodeErr != nil {
		return sizeEncodeErr
	}

	wLen, wErr := w.Write(c.sizeEnBytes[:])

	if wErr != nil {
		return wErr
	}

	if wLen != len(c.sizeEnBytes) {
		return ErrFailedFullyWriteSize
	}

	// Meta,
	// We use meta for Hash check
	hmac, hmacErr := c.hmac(c.writeMac, input)

	if hmacErr != nil {
		return hmacErr
	}

	wLen, wErr = w.Write(hmac)

	if wErr != nil {
		return wErr
	}

	if wLen != len(hmac) {
		return ErrFailedFullyWriteMeta
	}

	// Data
	wLen, wErr = w.Write(input)

	if wErr != nil {
		return wErr
	}

	if wLen != len(input) {
		return ErrFailedFullyWriteData
	}

	// Tail Padding
	tailPaddingErr := c.padding.WriteIn(w)

	if tailPaddingErr != nil {
		return tailPaddingErr
	}

	return nil
}

// Encode encrypt data and write result to the writer
func (c *crypter) Encode(input []byte, w io.Writer) error {
	c.encoder.XORKeyStream(input, input)

	wLen, wErr := w.Write(input)

	if wErr != nil {
		return wErr
	}

	if wLen != len(input) {
		return ErrFailedFullyWriteData
	}

	return nil
}

// Decode decrypt data and return the decrypted result
func (c *crypter) Decode(input []byte) ([]byte, error) {
	c.decoder.XORKeyStream(input, input)

	return input, nil
}

// Unstream parses decrypted data segment and return the meaningful data
func (c *crypter) Unstream(input io.Reader) (output []byte, err error) {
	// Font Padding
	fontPaddingErr := c.padding.Walkthrough(input)

	if fontPaddingErr != nil {
		return nil, fontPaddingErr
	}

	// Size
	rLen, rErr := input.Read(c.sizeDeBytes[:])

	if rErr != nil {
		return nil, rErr
	}

	if rLen != SizeSize {
		return nil, ErrFailedFullyReadSize
	}

	size := types.EncodableUint16(0)

	sizeDecodeErr := size.DecodeBytes(c.sizeDeBytes[:])

	if sizeDecodeErr != nil {
		return nil, sizeDecodeErr
	}

	if size <= 0 {
		return nil, ErrSizeTooSmall
	}

	if size > codec.BufferSize {
		return nil, ErrSizeTooLarge
	}

	// Meta
	rLen, rErr = input.Read(c.metaBytes[:])

	if rErr != nil {
		return nil, rErr
	}

	if rLen != MetaSize {
		return nil, ErrFailedFullyReadMeta
	}

	// Data
	rLen, rErr = input.Read(c.data[:size])

	if rErr != nil {
		return nil, rErr
	}

	if uint64(rLen) != uint64(size) {
		return nil, ErrFailedFullyReadData
	}

	// Tail Padding
	tailPaddingErr := c.padding.Walkthrough(input)

	if tailPaddingErr != nil {
		return nil, tailPaddingErr
	}

	// Check data
	dataHMAC, dataHMACErr := c.hmac(c.readMac, c.data[:size])

	if dataHMACErr != nil {
		return nil, dataHMACErr
	}

	if !c.hmacEqual(dataHMAC, c.metaBytes[:]) {
		return nil, ErrInvalidDataReaded
	}

	return c.data[:size], nil
}
