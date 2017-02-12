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

package main

import (
	"os"

	"github.com/nickrio/coward/application"
	"github.com/nickrio/coward/roles/channel"
	"github.com/nickrio/coward/roles/common/network/communicator/common/wrapper"
	"github.com/nickrio/coward/roles/proxy"
	"github.com/nickrio/coward/roles/socks5"
)

func main() {
	var execErr error

	app := application.New(application.Config{
		Banner:    "",
		Name:      "",
		Version:   "",
		Copyright: "",
		URL:       "",
		Components: application.Components{
			socks5.Role, proxy.Role, channel.Role,
			wrapper.Plain, wrapper.AESCFB128, wrapper.AESCFB256,
			wrapper.Chaotic,
		},
	})

	switch len(os.Args) {
	case 0:
	case 1:
		execErr = app.Help()
	default:
		execErr = app.ExecuteArgumentInput(os.Args[1:])
	}

	if execErr != nil {
		os.Exit(1)
	}
}
