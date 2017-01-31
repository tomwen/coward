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
	"fmt"
	"sync"
	"time"

	ccommon "github.com/nickrio/coward/common"
	"github.com/nickrio/coward/common/role"
	"github.com/nickrio/coward/roles/channel/common"
	"github.com/nickrio/coward/roles/channel/listener"
	"github.com/nickrio/coward/roles/common/network"
	"github.com/nickrio/coward/roles/common/network/transporter"
)

// channel implements Channel
type channel struct {
	transporter     transporter.Client
	cfg             Config
	defaultProc     ccommon.Proccessors
	protocols       common.Protocols
	servers         []common.ChannelServer
	serverIndex     int
	serverDownWait  sync.WaitGroup
	clientCloseWait sync.WaitGroup
	closeNotify     chan<- bool
}

// New builds a new Channel client
func New(transporter transporter.Client, cfg Config) role.Role {
	c := &channel{
		transporter:     transporter,
		cfg:             cfg,
		defaultProc:     network.GetDefaultProc(),
		protocols:       common.NewProtocols(),
		servers:         make([]common.ChannelServer, len(cfg.Channels)),
		serverIndex:     0,
		serverDownWait:  sync.WaitGroup{},
		clientCloseWait: sync.WaitGroup{},
		closeNotify:     nil,
	}

	c.protocols.Register(network.TCP, listener.NewTCP)
	c.protocols.Register(network.UDP, listener.NewUDP)

	return c
}

func (c *channel) Spawn(closeNotify chan<- bool) error {
	serverCloseWait := sync.WaitGroup{}

	c.serverDownWait.Add(1)

	// Wait for all server been shutdown
	go func() {
		keepCloseLoop := true

		serverCloseWait.Wait()

		go func() {
			defer c.serverDownWait.Done()

			for keepCloseLoop {
				for _, serv := range c.servers {
					serv.Server.Drop()
				}

				c.transporter.Kickoff()
			}
		}()

		c.clientCloseWait.Wait()

		keepCloseLoop = false
	}()

	for _, channel := range c.cfg.Channels {
		serverTimeout := channel.Timeout
		serverConcurrent := channel.Concurrence

		if serverTimeout == time.Duration(0) {
			serverTimeout = c.cfg.DefaultTimeout
		}

		if serverConcurrent == 0 {
			serverConcurrent = c.cfg.MaxConcurrence
		}

		server, serverCreationErr := c.protocols.CreateServer(
			channel.Protocol, common.ServerConfig{
				ID:          channel.ID,
				Interface:   c.cfg.Interface,
				Port:        channel.Port,
				Timeout:     serverTimeout,
				Concurrence: serverConcurrent,
				DefaultProc: c.defaultProc,
				Transporter: c.transporter,
				Logger: c.cfg.Logger.Context(fmt.Sprintf("[%s:%d] %s:%d",
					channel.Protocol.String(), channel.ID,
					c.cfg.Interface,
					channel.Port)),
			})

		if serverCreationErr != nil {
			c.cfg.Logger.Errorf(
				"Failed to create %s listener on %s:%d due to error: %s",
				channel.Protocol.String(), c.cfg.Interface,
				channel.Port, serverCreationErr)

			continue
		}

		c.servers[c.serverIndex] = common.ChannelServer{
			Protocol: channel.Protocol,
			Port:     channel.Port,
			Server:   server,
		}

		serverCloseWait.Add(1)

		go func(s common.ChannelServer) {
			defer func() {
				c.cfg.Logger.Infof("%s listener on %s:%d is down",
					s.Protocol.String(), c.cfg.Interface, s.Port)

				serverCloseWait.Done()
			}()

			c.cfg.Logger.Infof("Serving %s listener on %s:%d",
				s.Protocol.String(), c.cfg.Interface, s.Port)

			serveErr := s.Server.Serve(&c.clientCloseWait)

			if serveErr == nil {
				return
			}

			c.cfg.Logger.Debugf(
				"Failed to serve %s listener on %s:%d due to error: %s",
				s.Protocol.String(), c.cfg.Interface, s.Port, serveErr)
		}(c.servers[c.serverIndex])

		c.serverIndex++
	}

	c.closeNotify = closeNotify

	return nil
}

func (c *channel) Unspawn() error {
	for _, server := range c.servers[:c.serverIndex] {
		c.serverDownWait.Add(1)

		// Run in routine, or we will be blocked
		go func(s common.ChannelServer) {
			defer c.serverDownWait.Done()

			c.cfg.Logger.Debugf("Closing %s listener on %s:%d",
				s.Protocol.String(), c.cfg.Interface, s.Port)

			serverCloseErr := s.Server.Close(&c.clientCloseWait)

			if serverCloseErr == nil {
				c.cfg.Logger.Debugf("%s Listener on %s:%d is down",
					s.Protocol.String(), c.cfg.Interface, s.Port)

				return
			}

			c.cfg.Logger.Debugf(
				"Failed to close %s listener on %s:%d due to error: %s",
				s.Protocol.String(),
				c.cfg.Interface, s.Port, serverCloseErr)
		}(server)
	}

	c.serverDownWait.Wait()

	c.closeNotify <- true

	return nil
}
