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

package proxy

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/nickrio/coward/common/logger"
	"github.com/nickrio/coward/common/print"
	"github.com/nickrio/coward/common/role"
	"github.com/nickrio/coward/roles/common/network"
	ccommon "github.com/nickrio/coward/roles/common/network/communicator/common"
	"github.com/nickrio/coward/roles/common/network/communicator/common/wrapper"
	"github.com/nickrio/coward/roles/common/network/communicator/tcp"
	"github.com/nickrio/coward/roles/common/network/transporter"
	"github.com/nickrio/coward/roles/proxy/common"
)

// ConfigChannel is the bare configuration of proxy channel
type ConfigChannel struct {
	SelectProto network.Protocol
	ID          byte   `json:"id" cfg:"i,-id:Channel ID"`
	Host        string `json:"host" cfg:"h,-host:Target hostname"`
	Port        uint16 `json:"port" cfg:"p,-port:Target port"`
	Protocol    string `json:"protocol" cfg:"pr,-protocol:Protocol of the target connection"`
}

// VerifyHost verify Host field
func (c *ConfigChannel) VerifyHost() error {
	if c.Host == "" {
		return fmt.Errorf("Channel ID must not be empty")
	}

	return nil
}

// VerifyPort verify Port field
func (c *ConfigChannel) VerifyPort() error {
	if c.Port <= 0 {
		return fmt.Errorf("Channel Port must be greater than 0")
	}

	return nil
}

// VerifyProtocol verify Protocol field
func (c *ConfigChannel) VerifyProtocol() error {
	if c.Protocol == "" {
		return fmt.Errorf("Channel Protocol must not be empty")
	}

	protocol, protocolErr := network.GetProtocolByString(c.Protocol)

	if protocolErr != nil {
		return protocolErr
	}

	c.SelectProto = protocol

	return nil
}

// Verify verify current ConfigChannel object
func (c *ConfigChannel) Verify() error {
	if c.Host == "" {
		return fmt.Errorf("Channel Host must be defined")
	}

	if c.Port == 0 {
		return fmt.Errorf("Channel Port must be defined")
	}

	if c.Protocol == "" || c.SelectProto == network.Protocol(0) {
		return fmt.Errorf("Channel Protocol must be defined")
	}

	return nil
}

// ConfigInput is the bare configuration of proxy backend server
type ConfigInput struct {
	encryptAlgos        []wrapper.Wrapper
	encryptAlgosList    []string
	noisers             []wrapper.Disrupter
	noisersList         []string
	connPersistentSet   bool
	SelectedEncryptAlgo func(key []byte) ccommon.ConnWrapper
	SelectedNoiser      func(setting []byte) ccommon.ConnDisrupter
	SelectedChannels    common.Channels
	ListenIface         net.IP
	ListenAddr          string          `json:"listen_address" cfg:"la,-listen-address:Which address this backend server will listen on"`
	ListenPort          uint16          `json:"listen_port" cfg:"lp,-listen-port:Which port this backend server will listen on"`
	IdleTimeout         uint16          `json:"idle_timeout" cfg:"it,-idle-timeout:How long the connection can stay idle before been taken down"`
	ConnectTimeout      uint16          `json:"connection_timeout" cfg:"ct,-connection-timeout:The maximum wait time when we trying to establish a connection"`
	ConnPersistent      bool            `json:"connection_persistent" cfg:"cp,-connection-persistent:Whether or not to reuse idle connections for another request"`
	EncryptionAlgorithm string          `json:"encryption_algorithm" cfg:"ea,-encryption-algorithm:Which algorithm will be used to encrypt and obscure data"`
	EncryptionKey       string          `json:"encrypt_key" cfg:"ek,-encryption-key:Key (or Passphrase) for the encryption algorithm"`
	Noiser              string          `json:"noiser" cfg:"nr,-noiser:Which Disruptor will be used for decharacterization"`
	NoiserData          string          `json:"noiser_data" cfg:"nd,-noiser-data:Disruptor configuration string"`
	Channels            []ConfigChannel `json:"channels" cfg:"ch,-channels:Pre-defined destination"`
}

// GetDescription get additional information of a field
func (c ConfigInput) GetDescription(fieldPath string) string {
	result := ""

	switch fieldPath {
	case "/EncryptionAlgorithm":
		if len(c.encryptAlgosList) > 0 {
			result = "Available encryption algorithms are:\r\n- " +
				strings.Join(c.encryptAlgosList, "\r\n- ")
		}

	case "/Noiser":
		if len(c.noisersList) > 0 {
			result = "Available noisers are:\r\n- " +
				strings.Join(c.noisersList, "\r\n- ")
		}

	case "/Channels/Protocol":
		result = "Available protocols are:\r\n- " +
			strings.Join([]string{"tcp", "udp"}, "\r\n- ")
	}

	return result
}

// VerifyEncryptionAlgorithm verify EncryptionAlgorithm Field
func (c *ConfigInput) VerifyEncryptionAlgorithm() error {
	for _, algo := range c.encryptAlgos {
		if algo.Name != c.EncryptionAlgorithm {
			continue
		}

		c.SelectedEncryptAlgo = algo.Wrapper

		return nil
	}

	return fmt.Errorf("Encryption algorithm \"%s\" is undefined",
		c.EncryptionAlgorithm)
}

// VerifyNoiser verify Noiser Field
func (c *ConfigInput) VerifyNoiser() error {
	for _, noi := range c.noisers {
		if noi.Name != c.Noiser {
			continue
		}

		c.SelectedNoiser = noi.Disrupter

		return nil
	}

	return fmt.Errorf("Noiser \"%s\" is undefined", c.Noiser)
}

// VerifyListenPort verify ListenPort Field
func (c *ConfigInput) VerifyListenPort() error {
	if c.ListenPort <= 0 {
		return fmt.Errorf("Invalid Port number \"%d\"", c.ListenPort)
	}

	return nil
}

// VerifyListenAddr verify ListenAddr Field
func (c *ConfigInput) VerifyListenAddr() error {
	ipAddr := net.ParseIP(c.ListenAddr)

	if ipAddr == nil {
		return fmt.Errorf("Invalid IP address \"%s\"", c.ListenAddr)
	}

	c.ListenIface = ipAddr

	return nil
}

// VerifyIdleTimeout verify IdleTimeout Field
func (c *ConfigInput) VerifyIdleTimeout() error {
	if c.IdleTimeout <= 1 {
		return fmt.Errorf("Idle Timeout must greater than 1 second")
	}

	if c.IdleTimeout <= c.ConnectTimeout {
		return fmt.Errorf("Idle Timeout must greater than %d second",
			c.ConnectTimeout)
	}

	return nil
}

// VerifyConnectTimeout verify ConnectTimeout Field
func (c *ConfigInput) VerifyConnectTimeout() error {
	if c.ConnectTimeout <= 0 {
		return fmt.Errorf("Connection Timeout must greater than 0 second")
	}

	if c.ConnectTimeout >= c.IdleTimeout {
		return fmt.Errorf("Connection Timeout must smaller than "+
			"Idle Timeout (%d second)", c.IdleTimeout)
	}

	return nil
}

// VerifyConnPersistent verify ConnPersistent Field
func (c *ConfigInput) VerifyConnPersistent() error {
	c.connPersistentSet = true

	return nil
}

// VerifyEncryptionKey verify EncryptionKey Field
func (c *ConfigInput) VerifyEncryptionKey() error {
	if len(c.EncryptionKey) < 16 {
		return fmt.Errorf("Encryption Key must no shorter than 16 charactors")
	}

	return nil
}

// VerifyChannels verify Channels Field
func (c *ConfigInput) VerifyChannels() error {
	for _, channel := range c.Channels {
		selectedChannelErr := c.SelectedChannels.Add(common.Channel{
			ID:       channel.ID,
			Host:     channel.Host,
			Port:     channel.Port,
			Protocol: channel.SelectProto,
		})

		if selectedChannelErr != nil {
			return selectedChannelErr
		}
	}

	return nil
}

// Verify data ConfigInput after assign is completed
func (c *ConfigInput) Verify() error {
	if c.ListenAddr == "" {
		return errors.New("Listen Address must be defined")
	}

	if c.ListenPort <= 0 {
		return errors.New("Listen Port must be defined")
	}

	if c.IdleTimeout <= 0 {
		return errors.New("Idle Timeout must be defined")
	}

	if c.ConnectTimeout <= 0 {
		return errors.New("Connection Timeout must be defined")
	}

	if !c.connPersistentSet {
		c.ConnPersistent = true
	}

	if c.EncryptionAlgorithm == "" {
		return errors.New("Encryption Algorithm must be defined")
	}

	if c.EncryptionKey == "" {
		return errors.New("Encryption Key must be defined")
	}

	if c.Noiser == "" {
		return errors.New("Noiser must be defined")
	}

	return nil
}

// Role returns role registration information
func Role() role.Registration {
	return role.Registration{
		Name: "proxy",
		Description: "A proxy backend server which will handle and relay " +
			"incoming requests from a COWARD proxy client",
		Configurator: func(components role.Components) interface{} {
			encryptAlgos := []wrapper.Wrapper{}
			encryptAlgosList := []string{}
			noisers := []wrapper.Disrupter{}
			noisersList := []string{}

			for _, c := range components {
				switch component := c.(type) {
				case func() wrapper.Wrapper:
					cmp := component()

					encryptAlgos = append(encryptAlgos, cmp)
					encryptAlgosList = append(encryptAlgosList, cmp.Name)

				case func() wrapper.Disrupter:
					cmp := component()

					noisers = append(noisers, cmp)
					noisersList = append(noisersList, cmp.Name)
				}
			}

			return &ConfigInput{
				encryptAlgos:        encryptAlgos,
				encryptAlgosList:    encryptAlgosList,
				noisers:             noisers,
				noisersList:         noisersList,
				connPersistentSet:   false,
				SelectedEncryptAlgo: nil,
				SelectedNoiser:      nil,
				SelectedChannels:    common.Channels{},
				ListenIface:         net.ParseIP("127.0.0.1"),
				Channels:            []ConfigChannel{},
			}
		},
		Generater: func(
			w print.Common,
			config interface{},
			log logger.Logger,
		) (role.Role, error) {
			cfg := config.(*ConfigInput)

			tspServer := transporter.NewServer(tcp.NewServer(
				cfg.ListenIface,
				cfg.ListenPort,
				time.Duration(cfg.ConnectTimeout)*time.Second,
				time.Duration(cfg.IdleTimeout)*time.Second,
				cfg.SelectedEncryptAlgo([]byte(cfg.EncryptionKey)),
				cfg.SelectedNoiser([]byte(cfg.EncryptionKey)),
			), cfg.ConnPersistent)

			return New(tspServer, Config{
				Channels:       cfg.SelectedChannels,
				Logger:         log.Context("Proxy"),
				ConnectTimeout: time.Duration(cfg.ConnectTimeout) * time.Second,
				IdleTimeout:    time.Duration(cfg.IdleTimeout) * time.Second,
			}), nil
		},
	}
}
