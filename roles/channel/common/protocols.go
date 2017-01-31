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

import "github.com/nickrio/coward/roles/common/network"

// Protocols creates certain server according to registered protocols
type Protocols interface {
	Register(protocol network.Protocol, serverBuilder ServerBuild) error
	CreateServer(protocol network.Protocol, config ServerConfig) (Server, error)
}

// protocols implements Protocols
type protocols struct {
	protocols [network.MaxProtocolID]ServerBuild
}

// NewProtocols creates a new Protocols
func NewProtocols() Protocols {
	return &protocols{
		protocols: [network.MaxProtocolID]ServerBuild{},
	}
}

// Register registers a new protocol
func (p *protocols) Register(
	protocol network.Protocol, serverBuilder ServerBuild) error {
	if protocol >= network.MaxProtocolID {
		return ErrProtocolInvalidID
	}

	if p.protocols[protocol] != nil {
		return ErrProtocolAlreadyRegistered
	}

	p.protocols[protocol] = serverBuilder

	return nil
}

// CreateServer creates a server according specifed configuration
func (p *protocols) CreateServer(
	protocol network.Protocol, config ServerConfig) (Server, error) {
	if protocol >= network.MaxProtocolID {
		return nil, ErrProtocolInvalidID
	}

	if p.protocols[protocol] == nil {
		return nil, ErrProtocolUndefined
	}

	return p.protocols[protocol](config)
}
