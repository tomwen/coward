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

package messaging

import (
	"errors"
	"io"
	"math"
	"sync"

	"github.com/nickrio/coward/common"
	"github.com/nickrio/coward/common/types"
)

const (
	// HeadSize is the min size of parse buffer
	HeadSize = 3
)

var (
	// ErrParseBufferTooSmall is throwed when parse buffer is too
	// small
	ErrParseBufferTooSmall = errors.New(
		"Parse Buffer is too small")

	// ErrPrepareBufferTooSmall is throwed when write prepare buffer
	// is too small
	ErrPrepareBufferTooSmall = errors.New(
		"Write Prepare Buffer is too small")

	// ErrDataSizeTooLarge is throwed when the data size is too large
	ErrDataSizeTooLarge = errors.New(
		"Data size too large")
)

// Message format:
//
// +----+------+-------------+
// |CMD | SIZE | DATA STREAM |
// +----+------+-------------+
// | 1  |  2   |    SIZE     |
// +----+------+-------------+

// Messaging will send and recive data accroding to protocol
type Messaging struct {
	dispatchLock sync.Mutex
	receiveLock  sync.Mutex
}

// Dispatch parses data from read and execute a proccessor
// to proccess the data
func (m *Messaging) Dispatch(
	client io.ReadWriter, buf []byte, proc common.Proccessors) error {
	m.dispatchLock.Lock()

	defer m.dispatchLock.Unlock()

	bufLen := len(buf)

	if bufLen < HeadSize {
		return ErrParseBufferTooSmall
	}

	// Read command
	_, rErr := io.ReadFull(client, buf[:HeadSize])

	if rErr != nil {
		return rErr
	}

	cmd := common.Command(buf[0])

	// Decode size
	size := types.EncodableUint16(0)

	decodeErr := size.DecodeBytes(buf[1:3])

	if decodeErr != nil {
		return decodeErr
	}

	if int(size) > bufLen {
		return ErrDataSizeTooLarge
	}

	return proc.Execute(cmd, buf, client, uint16(size))
}

// Write writes data to writer with extra meta information
func (m *Messaging) Write(
	client io.Writer,
	cmd common.Command,
	data []byte,
	wpBuf []byte,
) (int, error) {
	m.receiveLock.Lock()

	defer m.receiveLock.Unlock()

	dataLen := len(data)
	toWrite := dataLen + 3

	if dataLen > math.MaxUint16 {
		return 0, ErrDataSizeTooLarge
	}

	if toWrite > len(wpBuf) {
		return 0, ErrPrepareBufferTooSmall
	}

	wpBuf[0] = byte(cmd)

	size := types.EncodableUint16(dataLen)

	encodeErr := size.EncodeBytes(wpBuf[1:3])

	if encodeErr != nil {
		return 0, encodeErr
	}

	// Data maybe nil
	if data != nil {
		copy(wpBuf[3:], data)
	}

	wDataLen, wDataErr := client.Write(wpBuf[:toWrite])

	// Adjust wDataLen to remove HeadSize from the count
	if wDataLen < HeadSize {
		wDataLen = HeadSize
	}

	if wDataErr != nil {
		return wDataLen - HeadSize, wDataErr
	}

	return wDataLen - HeadSize, nil
}
