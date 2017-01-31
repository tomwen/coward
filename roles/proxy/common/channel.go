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
	"math"
	"net"
	"strconv"

	"github.com/nickrio/coward/roles/common/network"
)

// Channel errors
var (
	ErrChannelInvalidID = errors.New(
		"Invalid channel ID")

	ErrChannelAlreadyExisted = errors.New(
		"Channel already existed")

	ErrChannelNotExisted = errors.New(
		"Channel not existed")
)

// Channel is proxy pre-defined destination
type Channel struct {
	ID       byte
	Host     string
	Port     uint16
	Protocol network.Protocol
}

// ChannelItem is the item of channel
type ChannelItem struct {
	Host     string
	Port     uint16
	Protocol network.Protocol
	Address  string
}

// Channels is the Channel Search List
type Channels [math.MaxUint8 + 1]*ChannelItem

// Add adds one item to the Channel Search List
func (c *Channels) Add(ch Channel) error {
	if ch.ID > math.MaxUint8 {
		return ErrChannelInvalidID
	}

	if c[ch.ID] != nil {
		return ErrChannelAlreadyExisted
	}

	c[ch.ID] = &ChannelItem{
		Host:     ch.Host,
		Port:     ch.Port,
		Protocol: ch.Protocol,
		Address: net.JoinHostPort(ch.Host, strconv.FormatUint(
			uint64(ch.Port), 10)),
	}

	return nil
}

// Get gets an item from Channel Search List
func (c *Channels) Get(id byte) (*ChannelItem, error) {
	if id > math.MaxUint8 {
		return nil, ErrChannelInvalidID
	}

	if c[id] == nil {
		return nil, ErrChannelNotExisted
	}

	return c[id], nil
}
