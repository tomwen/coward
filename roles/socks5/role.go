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

package socks5

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
	"github.com/nickrio/coward/roles/common/network/transporter"
	"github.com/nickrio/coward/roles/common/network/transporter/balancer"
	"github.com/nickrio/coward/roles/common/network/transporter/clients"
	tcomm "github.com/nickrio/coward/roles/common/network/transporter/common"
	"github.com/nickrio/coward/roles/common/network/wrapper"
)

// ConfigAuth is the bare configuration for --auth-user option
type ConfigAuth struct {
	User     string `json:"user" cfg:"u,-user:User name for login auth"`
	Password string `json:"password" cfg:"p,-pass:User password for login auth"`
}

// Verify checks ConfigAuth after assign is done
func (c ConfigAuth) Verify() error {
	if c.User == "" || c.Password == "" {
		return errors.New("Both the username and password must not be empty")
	}

	return nil
}

// EnAlgo is named string for EncryptionAlgorithm
type EnAlgo string

// Noiser is named string for Noiser
type Noiser string

// ConfigRemote is remote servers
type ConfigRemote struct {
	connPersistentSet   bool
	SelectedEncryptAlgo func(key []byte) tcomm.ConnWrapper
	SelectedNoiser      func(setting []byte) tcomm.ConnDisrupter
	RemoteHost          string `json:"remote_host" cfg:"rh,-host:Host name of the backend server"`
	RemotePort          uint16 `json:"remote_port" cfg:"rp,-port:Port of the backend server"`
	IdleTimeout         uint16 `json:"idle" cfg:"it,-idle:How long the connection can stay idle before been taken down"`
	ConnectTimeout      uint16 `json:"connection_timeout" cfg:"ct,-timeout:The maximum wait time when we trying to establish a connection"`
	ConnectRetry        uint8  `json:"connection_retry" cfg:"cr,-retry:How many times to retry when initial connection has failed"`
	ConnConcurrent      uint16 `json:"connection_concurrent" cfg:"cc,-concurrent:How many connections can be established with backend server at same time"`
	ConnPersistent      bool   `json:"connection_persistent" cfg:"cp,-persistent:Whether or not to reuse idle connections for another request"`
	EncryptionAlgorithm EnAlgo `json:"encryption_algorithm" cfg:"ea,-algorithm:Which algorithm will be used to encrypt and obscure data"`
	EncryptionKey       string `json:"encrypt_key" cfg:"ek,-key:Key (or Passphrase) for the encryption algorithm"`
	Noiser              Noiser `json:"noiser" cfg:"nr,-noiser:Which Disruptor will be used for decharacterization"`
	NoiserData          string `json:"noiser_data" cfg:"nd,-noiser-data:Disruptor configuration string"`
}

// VerifyRemoteHost Verify RemoteHost field
func (c *ConfigRemote) VerifyRemoteHost() error {
	if c.RemoteHost == "" {
		return fmt.Errorf("Invalid Host address \"%s\"", c.RemoteHost)
	}

	return nil
}

// VerifyRemotePort Verify RemotePort field
func (c *ConfigRemote) VerifyRemotePort() error {
	if c.RemotePort <= 0 {
		return fmt.Errorf("Invalid Port number \"%d\"", c.RemotePort)
	}

	return nil
}

// VerifyIdleTimeout Verify IdleTimeout field
func (c *ConfigRemote) VerifyIdleTimeout() error {
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
func (c *ConfigRemote) VerifyConnectTimeout() error {
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
func (c *ConfigRemote) VerifyConnectRetry() error {
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
func (c *ConfigRemote) VerifyConnConcurrent() error {
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
func (c *ConfigRemote) VerifyConnPersistent() error {
	c.connPersistentSet = true

	return nil
}

// VerifyEncryptionKey Verify EncryptionKey field
func (c *ConfigRemote) VerifyEncryptionKey() error {
	if len(c.EncryptionKey) < 16 {
		return fmt.Errorf("Encryption Key must no shorter than 16 charactors")
	}

	return nil
}

// Verify verifies ConfigRemote
func (c *ConfigRemote) Verify() error {
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

// ConfigInput is the bare configuration of socks5 server
type ConfigInput struct {
	encryptAlgos           []wrapper.Wrapper
	encryptAlgosList       []string
	noisers                []wrapper.Disrupter
	noisersList            []string
	ListenIface            net.IP
	AuthUsers              map[string]string
	Auth                   []ConfigAuth    `json:"auth_users" cfg:"au,-auth-users:User account and passwords of the Socks 5 server"`
	Remotes                []*ConfigRemote `json:"remotes" cfg:"rs,-remotes:Remote proxy backends"`
	ListenAddr             string          `json:"listen_address" cfg:"la,-listen-address:The interface which the Socks5 proxy server will listen on"`
	ListenPort             uint16          `json:"listen_port" cfg:"lp,-listen-port:The port which the Socks5 proxy server will listen on"`
	RememberedDestinations uint            `json:"remembered_destinations" cfg:"rd,-remembered-dest:How many destinations will be remembered for connection optimization"`
}

// GetDescription returns additional information about a field
func (c ConfigInput) GetDescription(fieldPath string) string {
	result := ""

	switch fieldPath {
	case "/Remotes/EncryptionAlgorithm":
		if len(c.encryptAlgosList) > 0 {
			result = "Available encryption algorithms are:\r\n- " +
				strings.Join(c.encryptAlgosList, "\r\n- ")
		}

	case "/Remotes/Noiser":
		if len(c.noisersList) > 0 {
			result = "Available noisers are:\r\n- " +
				strings.Join(c.noisersList, "\r\n- ")
		}
	}

	return result
}

// CheckValue checks items in a slice
func (c ConfigInput) CheckValue(name string, data interface{}) error {
	found := false

	switch d := data.(type) {
	case EnAlgo:
		for _, algo := range c.encryptAlgos {
			if algo.Name != string(d) {
				continue
			}

			found = true
		}

		if !found {
			return fmt.Errorf("Unknown Encryption Algorithm: %s", d)
		}

	case Noiser:
		for _, noi := range c.noisers {
			if noi.Name != string(d) {
				continue
			}

			found = true
		}

		if !found {
			return fmt.Errorf("Unknown Noiser: %s", d)
		}
	}

	return nil
}

// VerifyAuth Verify Auth field
func (c *ConfigInput) VerifyAuth() error {
	for _, a := range c.Auth {
		_, found := c.AuthUsers[a.User]

		if found {
			return fmt.Errorf("User \"%s\" defined twice", a.User)
		}

		c.AuthUsers[a.User] = a.Password
	}

	return nil
}

// VerifyRemotes Verify Remotes field
func (c *ConfigInput) VerifyRemotes() error {
	if len(c.Remotes) <= 0 {
		return errors.New("At least one remote must be defined")
	}

	for rIdx := range c.Remotes {
		for _, algo := range c.encryptAlgos {
			if algo.Name != string(c.Remotes[rIdx].EncryptionAlgorithm) {
				continue
			}

			c.Remotes[rIdx].SelectedEncryptAlgo = algo.Wrapper
		}

		if c.Remotes[rIdx].SelectedEncryptAlgo == nil {
			return fmt.Errorf("Unknown Encryption Algorithm: %s",
				c.Remotes[rIdx].EncryptionAlgorithm)
		}

		for _, noi := range c.noisers {
			if noi.Name != string(c.Remotes[rIdx].Noiser) {
				continue
			}

			c.Remotes[rIdx].SelectedNoiser = noi.Disrupter
		}

		if c.Remotes[rIdx].SelectedNoiser == nil {
			return fmt.Errorf("Unknown Noiser: %s",
				c.Remotes[rIdx].Noiser)
		}
	}

	return nil
}

// VerifyListenPort Verify ListenPort field
func (c *ConfigInput) VerifyListenPort() error {
	if c.ListenPort <= 0 {
		return fmt.Errorf("Invalid Port number \"%d\"", c.ListenPort)
	}

	return nil
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

// VerifyRememberedDestinations Verify Remembered Destinations field
func (c *ConfigInput) VerifyRememberedDestinations() error {
	if c.RememberedDestinations <= 0 {
		return errors.New("Remembered Destinations must be greater than 0")
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

	if c.ListenPort <= 0 {
		c.ListenPort = 1080
	}

	if c.RememberedDestinations <= 0 {
		c.RememberedDestinations = 1024
	}

	if len(c.Remotes) <= 0 {
		return errors.New("Remote must be defined")
	}

	return nil
}

// Role returns role registration information
func Role() role.Registration {
	return role.Registration{
		Name: "socks5",
		Description: "Pretend to be a Socks 5 proxy server and redirect all " +
			"recevied requests to remote backend servers",
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
				encryptAlgos:     encryptAlgos,
				encryptAlgosList: encryptAlgosList,
				noisers:          noisers,
				noisersList:      noisersList,
				ListenIface:      net.ParseIP("127.0.0.1"),
				AuthUsers:        map[string]string{},
				Remotes:          []*ConfigRemote{},
			}
		},
		Generater: func(
			w print.Common,
			config interface{},
			log logger.Logger,
		) (role.Role, error) {
			var auther func(user string, pass string) error
			var maxIdleDuration uint16

			cfg := config.(*ConfigInput)

			if len(cfg.AuthUsers) > 0 {
				auther = func(user string, pass string) error {
					u, has := cfg.AuthUsers[user]

					if !has {
						return fmt.Errorf("User \"%s\" was not found", user)
					}

					if u != pass {
						return fmt.Errorf(
							"User \"%s\" login with a wrong password", user)
					}

					return nil
				}
			}

			transporters := make([]transporter.Client, len(cfg.Remotes))

			for transportIndex, transportCfg := range cfg.Remotes {
				transporters[transportIndex] = transporter.NewClient(
					transportCfg.RemoteHost,
					transportCfg.RemotePort,
					time.Duration(transportCfg.IdleTimeout)*time.Second,
					transportCfg.ConnPersistent,
					transportCfg.ConnConcurrent,
					transportCfg.ConnectRetry,
					time.Duration(transportCfg.ConnectTimeout)*time.Second,
					transportCfg.SelectedEncryptAlgo(
						[]byte(transportCfg.EncryptionKey)),
					transportCfg.SelectedNoiser(
						[]byte(transportCfg.NoiserData)))

				if transportCfg.IdleTimeout > maxIdleDuration {
					maxIdleDuration = transportCfg.IdleTimeout
				}
			}

			return New(
				balancer.New(
					clients.New(transporters), cfg.RememberedDestinations),
				Config{
					Auth:      auther,
					Timeout:   time.Duration(maxIdleDuration) * time.Second,
					Interface: cfg.ListenIface,
					Port:      cfg.ListenPort,
					Logger:    log.Context("Socks5"),
				}), nil
		},
	}
}
