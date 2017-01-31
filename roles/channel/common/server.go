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
	"net"
	"sync"
	"time"

	"github.com/nickrio/coward/common"
	"github.com/nickrio/coward/common/logger"
	"github.com/nickrio/coward/roles/common/network/transporter"
)

// ServerConfig is the server config
type ServerConfig struct {
	ID          byte
	Interface   net.IP
	Port        uint16
	Timeout     time.Duration
	Concurrence uint16
	DefaultProc common.Proccessors
	Transporter transporter.Client
	Logger      logger.Logger
}

// Server repersents a server
type Server interface {
	Serve(clientCloseWait *sync.WaitGroup) error
	Drop()
	Close(clientCloseWait *sync.WaitGroup) error
}

// ServerBuild is a function what will build a Server according
// input ServerConfig
type ServerBuild func(ServerConfig) (Server, error)
