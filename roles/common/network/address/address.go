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

package address

import (
	"errors"

	"github.com/nickrio/coward/common/types"
)

// Type is the Address Type
type Type byte

// Address Types
const (
	Unknown Type = 0
	UseLast Type = 1
	IPv4    Type = 2
	IPv6    Type = 3
	Domain  Type = 4
)

// Lengths
const (
	ipv4AddrLen           = 4
	ipv4AddrPortLen       = 6
	ipv4AddrPortResultLen = 7
	ipv6AddrLen           = 16
	ipv6AddrPortLen       = 18
	ipv6AddrPortResultLen = 19
	domainMinLen          = 1
	domainMaxLen          = 255
)

// Address errors
var (
	ErrUnsupportedAddress = errors.New(
		"Unsupported address")

	ErrUnknownAddressType = errors.New(
		"Unknown address type")

	ErrAddressDataTooShort = errors.New(
		"Unknown data too short")

	ErrInvalidAddressData = errors.New(
		"Invalid address data")

	ErrInvalidIPv4Length = errors.New(
		"Invalid IPv4 data length")

	ErrInvalidIPv4BufferLength = errors.New(
		"Invalid IPv4 data buffer length")

	ErrInvalidIPv6Length = errors.New(
		"Invalid IPv6 data length")

	ErrInvalidIPv6BufferLength = errors.New(
		"Invalid IPv6 data buffer length")

	ErrInvalidDomainLength = errors.New(
		"Invalid Domain data length")

	ErrInvalidDomainBufferLength = errors.New(
		"Invalid Domain data buffer length")

	ErrInvalidUseLastBufferLength = errors.New(
		"Invalid UseLast data buffer length")
)

// Default is the default address parser
var Default = Address{}

// Address is the address encoder
type Address struct{}

// Pack packs address accroding to Address Type
func (a *Address) Pack(
	aType Type, addr []byte, port uint16, result []byte) (int, error) {
	switch aType {
	case UseLast:
		return a.packLast(addr, port, result)

	case IPv4:
		return a.packIPv4(addr, port, result)

	case IPv6:
		return a.packIPv6(addr, port, result)

	case Domain:
		return a.packDomain(addr, port, result)
	}

	return 0, ErrUnknownAddressType
}

// Unpack unpacks bytes and extract address data from it
func (a *Address) Unpack(data []byte) (Type, []byte, uint16, int, error) {
	var address []byte
	var port types.EncodableUint16
	var dataConsumed int
	var err error

	dataLen := len(data)

	if dataLen < 1 {
		return Unknown, nil, 0, 0, ErrInvalidAddressData
	}

	addressType := Type(data[0])

	switch addressType {
	case UseLast:
		address, port, dataConsumed, err = a.unpackLast(data[1:], dataLen-1)

	case IPv4:
		address, port, dataConsumed, err = a.unpackIPv4(data[1:], dataLen-1)

	case IPv6:
		address, port, dataConsumed, err = a.unpackIPv6(data[1:], dataLen-1)

	case Domain:
		address, port, dataConsumed, err = a.unpackDomain(data[1:], dataLen-1)

	default:
		return Unknown, nil, 0, 0, ErrUnknownAddressType
	}

	if err != nil {
		return Unknown, nil, 0, 0, err
	}

	return addressType, address, uint16(port), dataConsumed + 1, err
}

// packLast build a empty address information
func (a *Address) packLast(l []byte, port uint16, result []byte) (int, error) {
	if len(result) < 1 {
		return 0, ErrInvalidUseLastBufferLength
	}

	// Mark it as Domain
	result[0] = byte(UseLast)

	return 1, nil
}

// pack takes an IPv4 address and a port and convert it to bytes.
func (a *Address) packIPv4(ip []byte, port uint16, result []byte) (int, error) {
	resultLen := len(result)

	if len(ip) != ipv4AddrLen {
		return 0, ErrInvalidIPv4Length
	}

	if resultLen < ipv4AddrPortResultLen {
		return 0, ErrInvalidIPv4BufferLength
	}

	// Mark it as IPv4
	result[0] = byte(IPv4)

	// Copy the first 4 bytes as the address
	copy(result[1:5], ip[:ipv4AddrLen])

	// Encode port to bytes
	portEncoder := types.EncodableUint16(port)

	portEncodeErr := portEncoder.EncodeBytes(result[5:ipv4AddrPortResultLen])

	if portEncodeErr != nil {
		return 0, portEncodeErr
	}

	return ipv4AddrPortResultLen, nil
}

// pack takes an IPv6 address and a port and convert it to bytes.
func (a *Address) packIPv6(ip []byte, port uint16, result []byte) (int, error) {
	resultLen := len(result)

	if len(ip) != ipv6AddrLen {
		return 0, ErrInvalidIPv6Length
	}

	if resultLen < ipv6AddrPortResultLen {
		return 0, ErrInvalidIPv6BufferLength
	}

	// Mark it as IPv6
	result[0] = byte(IPv6)

	// Copy the first 16 bytes as the address
	copy(result[1:17], ip[:ipv6AddrLen])

	// Encode port to bytes
	portEncoder := types.EncodableUint16(port)

	portEncodeErr := portEncoder.EncodeBytes(result[17:ipv6AddrPortResultLen])

	if portEncodeErr != nil {
		return 0, portEncodeErr
	}

	return ipv6AddrPortResultLen, nil
}

func (a *Address) packDomain(
	domain []byte, port uint16, result []byte) (int, error) {
	resultLen := len(result)
	domainLen := len(domain)
	end := domainLen + 4 // Type + Length + Port = 4 bytes

	if domainLen < domainMinLen || domainLen > domainMaxLen {
		return 0, ErrInvalidDomainLength
	}

	if resultLen < end {
		return 0, ErrInvalidDomainBufferLength
	}

	// Mark it as Domain
	result[0] = byte(Domain)

	// Record domain length
	result[1] = byte(domainLen)

	// Copy the first 16 bytes as the address
	copy(result[2:], domain[:domainLen])

	// Encode port to bytes
	portEncoder := types.EncodableUint16(port)

	portEncodeErr := portEncoder.EncodeBytes(result[domainLen+2 : end])

	if portEncodeErr != nil {
		return 0, portEncodeErr
	}

	return end, nil
}

// unpackIPv4 build a empty address from byte
func (a *Address) unpackLast(
	data []byte, length int) ([]byte, types.EncodableUint16, int, error) {
	return nil, 0, 0, nil
}

// unpackIPv4 unpacks and extract IPv4 address from data bytes
func (a *Address) unpackIPv4(
	data []byte, length int) ([]byte, types.EncodableUint16, int, error) {
	if length < ipv4AddrPortLen {
		return nil, 0, 0, ErrInvalidIPv4Length
	}

	addr := data[:ipv4AddrLen]

	port := types.EncodableUint16(0)

	portDecodeErr := port.DecodeBytes(data[ipv4AddrLen:ipv4AddrPortLen])

	if portDecodeErr != nil {
		return nil, 0, 0, portDecodeErr
	}

	return addr, port, ipv4AddrPortLen, nil
}

// unpackIPv6 unpacks and extract IPv6 address from data bytes
func (a *Address) unpackIPv6(
	data []byte, length int) ([]byte, types.EncodableUint16, int, error) {
	if length < ipv6AddrPortLen {
		return nil, 0, 0, ErrInvalidIPv6Length
	}

	addr := data[:ipv6AddrLen]

	port := types.EncodableUint16(0)

	portDecodeErr := port.DecodeBytes(data[ipv6AddrLen:ipv6AddrPortLen])

	if portDecodeErr != nil {
		return nil, 0, 0, portDecodeErr
	}

	return addr, port, ipv6AddrPortLen, nil
}

// unpackDomain unpacks and extract Domain address from data bytes
func (a *Address) unpackDomain(
	data []byte, length int) ([]byte, types.EncodableUint16, int, error) {
	if length < 1 {
		return nil, 0, 0, ErrInvalidDomainLength
	}

	domainLen := int(data[0])

	if length-1 < domainLen {
		return nil, 0, 0, ErrInvalidDomainLength
	}

	addr := data[1 : domainLen+1]

	port := types.EncodableUint16(0)

	portDecodeErr := port.DecodeBytes(data[domainLen+1 : domainLen+3])

	if portDecodeErr != nil {
		return nil, 0, 0, portDecodeErr
	}

	return addr, port, domainLen + 3, nil
}
