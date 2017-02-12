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

package common

import (
	"errors"
	"net"

	"github.com/nickrio/coward/common/types"
)

// Address is the address parser
type Address struct{}

// Address errors
var (
	ErrUnknownAddressType = errors.New(
		"Unknown Socks5 address type")

	ErrUnknownAddressData = errors.New(
		"Unknown Socks5 address data")

	ErrInvalidIPv4DataLength = errors.New(
		"Invalid IPv4 data length for the Socks5 request")

	ErrInvalidIPv4BufferLength = errors.New(
		"Invalid IPv4 buffer length for the Socks5 request")

	ErrInvalidIPv6DataLength = errors.New(
		"Invalid IPv6 data length for the Socks5 request")

	ErrInvalidIPv6BufferLength = errors.New(
		"Invalid IPv6 buffer length for the Socks5 request")

	ErrInvalidDomainData = errors.New(
		"Invalid Domain data for the Socks5 request")

	ErrInvalidDomainBufferLength = errors.New(
		"Invalid Domain buffer length for the Socks5 request")
)

const (
	ipv4AddrLen         = 4
	ipv4AddrPortLen     = 6
	ipv4TypeAddrPortLen = 7
	ipv6AddrLen         = 16
	ipv6AddrPortLen     = 18
	ipv6TypeAddPortLen  = 19
	domainMaxLength     = 255
)

// Pack packs host port data in to a string of bytes
func (a *Address) Pack(
	t ATYPE, addr []byte, port uint16, result []byte) (int, error) {
	switch t {
	case IPv4:
		return a.packIPv4(addr, types.EncodableUint16(port), result)

	case IPv6:
		return a.packIPv6(addr, types.EncodableUint16(port), result)

	case Domain:
		return a.packDomain(addr, types.EncodableUint16(port), result)
	}

	return 0, ErrUnknownAddressType
}

// PackIP packs an IP address to IPv6 or IPv4
func (a *Address) PackIP(ip net.IP, port uint16, result []byte) (int, error) {
	ipv4 := ip.To4()

	if ipv4 != nil {
		return a.Pack(IPv4, ipv4, port, result)
	}

	ipv6 := ip.To16()

	if ipv6 != nil {
		return a.Pack(IPv6, ipv6, port, result)
	}

	return 0, ErrUnknownAddressType
}

// Unpack extract address information out of given bytes
func (a *Address) Unpack(data []byte) (ATYPE, []byte, uint16, int, error) {
	var address []byte
	var port types.EncodableUint16
	var dataConsumed int
	var err error

	dataLen := len(data)

	if dataLen < 1 {
		return Unknown, nil, 0, 0, ErrUnknownAddressData
	}

	atype := ATYPE(data[0])

	switch atype {
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

	return atype, address, uint16(port), dataConsumed + 1, err
}

func (a *Address) unpackIPv4(
	b []byte, bLen int) ([]byte, types.EncodableUint16, int, error) {
	if bLen < ipv4AddrPortLen {
		return nil, 0, 0, ErrInvalidIPv4DataLength
	}

	addr := b[:ipv4AddrLen]

	port := types.EncodableUint16(0)

	decodeErr := port.DecodeBytes(b[ipv4AddrLen:ipv4AddrPortLen])

	if decodeErr != nil {
		return nil, 0, 0, decodeErr
	}

	return addr, port, ipv4AddrPortLen, nil
}

func (a *Address) packIPv4(
	addr []byte, port types.EncodableUint16, result []byte) (int, error) {
	if len(addr) != ipv4AddrLen {
		return 0, ErrInvalidIPv4DataLength
	}

	if len(result) < ipv4TypeAddrPortLen {
		return 0, ErrInvalidIPv4BufferLength
	}

	result[0] = byte(IPv4)

	copy(result[1:5], addr[:ipv4AddrLen])

	portEncodeErr := port.EncodeBytes(result[5:ipv4TypeAddrPortLen])

	if portEncodeErr != nil {
		return 0, portEncodeErr
	}

	return ipv4TypeAddrPortLen, nil
}

func (a *Address) unpackIPv6(
	b []byte, bLen int) ([]byte, types.EncodableUint16, int, error) {
	if bLen < ipv6AddrPortLen {
		return nil, 0, 0, ErrInvalidIPv6DataLength
	}

	addr := b[:ipv6AddrLen]

	port := types.EncodableUint16(0)

	decodeErr := port.DecodeBytes(b[ipv6AddrLen:ipv6AddrPortLen])

	if decodeErr != nil {
		return nil, 0, 0, decodeErr
	}

	return addr, port, ipv6AddrPortLen, nil
}

func (a *Address) packIPv6(
	addr []byte, port types.EncodableUint16, result []byte) (int, error) {
	if len(addr) != ipv6AddrLen {
		return 0, ErrInvalidIPv6DataLength
	}

	if len(result) < ipv6TypeAddPortLen {
		return 0, ErrInvalidIPv6BufferLength
	}

	result[0] = byte(IPv6)

	copy(result[1:17], addr[:ipv6AddrLen])

	portEncodeErr := port.EncodeBytes(result[17:ipv6TypeAddPortLen])

	if portEncodeErr != nil {
		return 0, portEncodeErr
	}

	return ipv6TypeAddPortLen, nil
}

func (a *Address) unpackDomain(
	b []byte, bLen int) ([]byte, types.EncodableUint16, int, error) {
	if bLen < 1 {
		return nil, 0, 0, ErrInvalidDomainData
	}

	domainLen := int(b[0])
	end := domainLen + 2

	if bLen < end {
		return nil, 0, 0, ErrInvalidDomainData
	}

	addr := b[1 : domainLen+1]

	port := types.EncodableUint16(0)

	decodeErr := port.DecodeBytes(b[domainLen+1 : domainLen+3])

	if decodeErr != nil {
		return nil, 0, 0, decodeErr
	}

	return addr, port, domainLen + 3, nil
}

func (a *Address) packDomain(
	addr []byte, port types.EncodableUint16, result []byte) (int, error) {
	addrLen := len(addr)
	resultLen := len(result)

	if addrLen < 1 {
		return 0, ErrInvalidDomainData
	}

	if addrLen > domainMaxLength {
		return 0, ErrInvalidDomainData
	}

	end := addrLen + 4

	if end > resultLen {
		return 0, ErrInvalidDomainBufferLength
	}

	result[0] = byte(Domain)
	result[1] = byte(addrLen)

	copy(result[2:], addr[:addrLen])

	portEncodeErr := port.EncodeBytes(result[addrLen+2 : end])

	if portEncodeErr != nil {
		return 0, portEncodeErr
	}

	return end, nil
}
