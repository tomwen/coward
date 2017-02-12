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
	"errors"
	"fmt"
	"math"
	"net"
	"strings"
	"time"

	"github.com/nickrio/coward/common/logger"
	"github.com/nickrio/coward/common/print"
	"github.com/nickrio/coward/common/role"
	"github.com/nickrio/coward/roles/common/network"
	ccomm "github.com/nickrio/coward/roles/common/network/communicator/common"
	"github.com/nickrio/coward/roles/common/network/communicator/common/wrapper"
	"github.com/nickrio/coward/roles/common/network/communicator/tcp"
	"github.com/nickrio/coward/roles/common/network/transporter"
)

// ConfigChannel is the bare channel setting input
type ConfigChannel struct {
	SelectedProtocol network.Protocol
	ID               byte   `json:"id" cfg:"i,-id:Server Channel ID, must be configured according to server setting"`
	Port             uint16 `json:"port" cfg:"p,-port:Local service port, all data sent to this port will be relayed to remote server"`
	Protocol         string `json:"protocol" cfg:"pt,-protocol:Protocol of the local port, must be configured according to server setting"`
	Timeout          uint16 `json:"timeout" cfg:"to,-timeout:Timeout of the local port"`
	Concurrent       uint16 `json:"concurrence" cfg:"cc,-concurrence:maximum concurrence of the local port"`
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

	c.SelectedProtocol = protocol

	return nil
}

// Verify verify current ConfigChannel object
func (c ConfigChannel) Verify() error {
	if c.Port == 0 {
		return fmt.Errorf("Channel Port must be defined")
	}

	if c.Protocol == "" || c.SelectedProtocol == network.Protocol(0) {
		return fmt.Errorf("Channel Protocol must be defined")
	}

	return nil
}

// ConfigInput is the bare configuration of channel client
type ConfigInput struct {
	encryptAlgos        []wrapper.Wrapper
	encryptAlgosList    []string
	noisers             []wrapper.Disrupter
	noisersList         []string
	connPersistentSet   bool
	ListenIface         net.IP
	SelectedEncryptAlgo func(key []byte) ccomm.ConnWrapper
	SelectedNoiser      func(setting []byte) ccomm.ConnDisrupter
	Channels            []ConfigChannel `json:"channels" cfg:"ch,-channels:Channels, must be configured according to server setting"`
	ListenAddr          string          `json:"local_address" cfg:"la,-listen-address:Which interface (Local IP address) this Channel client will serve on"`
	RemoteHost          string          `json:"remote_host" cfg:"rh,-remote-host:Host name of the backend server"`
	RemotePort          uint16          `json:"remote_port" cfg:"rp,-remote-port:Port of the backend server"`
	IdleTimeout         int64           `json:"idle_timeout" cfg:"it,-idle-timeout:How long the connection can stay idle before been taken down"`
	ConnectTimeout      int64           `json:"connection_timeout" cfg:"ct,-connection-timeout:The maximum wait time when we trying to establish a connection"`
	ConnectRetry        uint8           `json:"connection_retry" cfg:"cr,-connection-retry:How many times to retry when inital connection has failed"`
	ConnConcurrent      uint16          `json:"connection_concurrent" cfg:"cc,-connection-concurrent:How many connections can be established with backend server at same time"`
	ConnPersistent      bool            `json:"connection_persistent" cfg:"cp,-connection-persistent:Whether or not to reuse idle connections for another request"`
	EncryptionAlgorithm string          `json:"encryption_algorithm" cfg:"ea,-encryption-algorithm:Which algorithm will be used to encrypt and obscure data"`
	EncryptionKey       string          `json:"encrypt_key" cfg:"ek,-encryption-key:Key (or Passphrase) for the encryption algorithm"`
	Noiser              string          `json:"noiser" cfg:"nr,-noiser:Which Disruptor will be used for decharacterization"`
	NoiserData          string          `json:"noiser_data" cfg:"nd,-noiser-data:Disruptor configuration string"`
}

// GetDescription returns additional information about a field
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

// VerifyEncryptionAlgorithm Verify EncryptionAlgorithm field
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

// VerifyNoiser Verify Noiser field
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

// VerifyListenAddr Verify ListenAddr field
func (c *ConfigInput) VerifyListenAddr() error {
	ipAddr := net.ParseIP(c.ListenAddr)

	if ipAddr == nil {
		return fmt.Errorf("Invalid IP address \"%s\"", c.ListenAddr)
	}

	c.ListenIface = ipAddr

	return nil
}

// VerifyRemoteHost Verify RemoteHost field
func (c *ConfigInput) VerifyRemoteHost() error {
	if c.RemoteHost == "" {
		return fmt.Errorf("Invalid Host address \"%s\"", c.RemoteHost)
	}

	return nil
}

// VerifyRemotePort Verify RemotePort field
func (c *ConfigInput) VerifyRemotePort() error {
	if c.RemotePort <= 0 {
		return fmt.Errorf("Invalid Port number \"%d\"", c.RemotePort)
	}

	return nil
}

// VerifyIdleTimeout Verify IdleTimeout field
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

// VerifyConnectTimeout Verify ConnectTimeout field
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

// VerifyConnectRetry Verify ConnectRetry field
func (c *ConfigInput) VerifyConnectRetry() error {
	if c.ConnectRetry <= 0 {
		return fmt.Errorf("Connection Retry must greater than 0")
	}

	if c.ConnectRetry > math.MaxUint8 {
		return fmt.Errorf("Connection Retry must be smaller than %d",
			math.MaxUint8)
	}

	return nil
}

// VerifyConnConcurrent Verify ConnConcurrent field
func (c *ConfigInput) VerifyConnConcurrent() error {
	if c.ConnConcurrent <= 0 {
		return fmt.Errorf("Connection Concurrent must greater than 0")
	}

	if c.ConnConcurrent > math.MaxUint16 {
		return fmt.Errorf("Connection Concurrent must be smaller than %d",
			math.MaxUint16)
	}

	return nil
}

// VerifyConnPersistent Verify ConnPersistent field
func (c *ConfigInput) VerifyConnPersistent() error {
	c.connPersistentSet = true

	return nil
}

// VerifyEncryptionKey Verify EncryptionKey field
func (c *ConfigInput) VerifyEncryptionKey() error {
	if len(c.EncryptionKey) < 16 {
		return fmt.Errorf("Encryption Key must no shorter than 16 charactors")
	}

	return nil
}

// Verify checks ConfigInput after assign is completed
func (c *ConfigInput) Verify() error {
	if c.ListenAddr == "" {
		c.ListenAddr = "127.0.0.1"

		verifyErr := c.VerifyListenAddr()

		if verifyErr != nil {
			return verifyErr
		}
	}

	if c.RemoteHost == "" {
		return errors.New("Remote Host must be defined")
	}

	if c.RemotePort <= 0 {
		return errors.New("Remote Port must be defined")
	}

	if c.IdleTimeout <= 0 {
		return errors.New("Idle Timeout must be defined")
	}

	if c.ConnectTimeout <= 0 {
		return errors.New("Connection Timeout must be defined")
	}

	if c.ConnectRetry <= 0 {
		return errors.New("Connection Retry must be defined")
	}

	if c.ConnConcurrent <= 0 {
		return errors.New("Connection Concurrent must be defined")
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
		Name: "channel",
		Description: "Map pre-defined ports on the remote backend server" +
			" to local client",
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

			channels := make([]Channel, len(cfg.Channels))

			for idx, ch := range cfg.Channels {
				channels[idx] = Channel{
					ID:          ch.ID,
					Port:        ch.Port,
					Protocol:    ch.SelectedProtocol,
					Timeout:     time.Duration(ch.Timeout) * time.Second,
					Concurrence: ch.Concurrent,
				}
			}

			transport := transporter.NewClient(
				tcp.NewClientBuilder(
					cfg.RemoteHost,
					cfg.RemotePort,
					time.Duration(cfg.ConnectTimeout)*time.Second,
					time.Duration(cfg.IdleTimeout)*time.Second,
					cfg.SelectedEncryptAlgo(
						[]byte(cfg.EncryptionKey)),
					cfg.SelectedNoiser(
						[]byte(cfg.NoiserData)),
				),
				time.Duration(cfg.ConnectTimeout)*time.Second,
				cfg.ConnConcurrent,
				cfg.ConnectRetry,
				cfg.ConnPersistent,
			)

			return New(transport, Config{
				DefaultTimeout: time.Duration(cfg.ConnectTimeout) * time.Second,
				MaxConcurrence: cfg.ConnConcurrent,
				Interface:      cfg.ListenIface,
				Logger:         log.Context("Channel"),
				Channels:       channels,
			}), nil
		},
	}
}
