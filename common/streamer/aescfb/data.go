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

import "github.com/nickrio/coward/common/codec"

const (
	// HeadSize is the byte length of head data (IV in this case)
	HeadSize = 16

	// PaddingMaxSize is maximum padding size: Padding length
	// cannot larger than this
	PaddingMaxSize = 16

	// MetaSize is the byte length of meta data (HMAC in this case)
	MetaSize = 10

	// SizeSize is the byte length of the data size (an uint16 data)
	SizeSize = 2
)

var (
	// ErrFailedFullyReadHead is throwed when Head is not fully
	// readed
	ErrFailedFullyReadHead = codec.Fail(
		"Failed to fully read head")

	// ErrInvalidDataReaded is throwed when readed data is invalid
	ErrInvalidDataReaded = codec.Fail(
		"Invalid data readed")

	// ErrFailedFullyWriteHead is throwed when Head is not fully
	// written
	ErrFailedFullyWriteHead = codec.Fail(
		"Failed to fully write head")

	// ErrFailedFullyWriteFontPadding is throwed when Font Padding
	//  is not fully written
	ErrFailedFullyWriteFontPadding = codec.Fail(
		"Failed to fully write font padding")

	// ErrFailedFullyReadFontPadding is throwed when Font Padding
	//  is not fully readed
	ErrFailedFullyReadFontPadding = codec.Fail(
		"Failed to fully read font padding")

	// ErrFailedFullyWriteTailPadding is throwed when Tail Padding
	//  is not fully written
	ErrFailedFullyWriteTailPadding = codec.Fail(
		"Failed to fully write tail padding")

	// ErrFailedFullyReadTailPadding is throwed when Tail Padding
	//  is not fully readed
	ErrFailedFullyReadTailPadding = codec.Fail(
		"Failed to fully read tail padding")

	// ErrFailedFullyWriteMeta is throwed when Meta is not fully
	// written
	ErrFailedFullyWriteMeta = codec.Fail(
		"Failed to fully write meta")

	// ErrFailedFullyReadMeta is throwed when Meta is not fully
	// readed
	ErrFailedFullyReadMeta = codec.Fail(
		"Failed to fully write meta")

	// ErrFailedFullyWriteSize is throwed when Size is not fully
	// written
	ErrFailedFullyWriteSize = codec.Fail(
		"Failed to fully write size")

	// ErrFailedFullyReadSize is throwed when Size is not fully
	// readed
	ErrFailedFullyReadSize = codec.Fail(
		"Failed to fully read size")

	// ErrSizeTooLarge is throwed when Size is abnormal
	ErrSizeTooLarge = codec.Fail(
		"Data length is excceed the limit")

	// ErrSizeTooSmall is throwed when Size is 0
	ErrSizeTooSmall = codec.Fail(
		"Data length is too small")

	// ErrFailedFullyWriteData is throwed when Data is not fully
	// written
	ErrFailedFullyWriteData = codec.Fail(
		"Failed to fully write data")

	// ErrFailedFullyReadData is throwed when Data is not fully
	// readed
	ErrFailedFullyReadData = codec.Fail(
		"Failed to fully read data")

	// ErrFailedToWriteHMACToSumer is throwed when HMAC is not fully
	// written to summer
	ErrFailedToWriteHMACToSumer = codec.Fail(
		"Failed to write HMAC to summer")

	// ErrFailedHMACTooShort is throwed when HMAC bytes is too short
	ErrFailedHMACTooShort = codec.Fail(
		"HMAC is too short")
)
