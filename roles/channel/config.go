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

package channel

import (
	"net"
	"time"

	"github.com/nickrio/coward/common/logger"
	"github.com/nickrio/coward/roles/common/network"
)

// Channel is the local ports what will relay request to the remote server
type Channel struct {
	ID          byte
	Port        uint16
	Protocol    network.Protocol
	Timeout     time.Duration
	Concurrence uint16
}

// Config is the channel configuration
type Config struct {
	DefaultTimeout time.Duration
	MaxConcurrence uint16
	Interface      net.IP
	Logger         logger.Logger
	Channels       []Channel
}
