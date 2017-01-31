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

package udp

import (
	"errors"
	"time"
)

// Nodes errors
var (
	ErrNodeAlreadyExisted = errors.New(
		"Node already existed")

	ErrNodeNotFound = errors.New(
		"Node not found")
)

type nodesConfig struct {
	Size     uint16
	Expire   time.Duration
	OnDelete func(udpClient)
}

type nodeSearch map[string]*node

type node struct {
	name     string
	next     *node
	previous *node
	lastUse  time.Time
	conn     udpClient
}

type nodes struct {
	config  nodesConfig
	nodes   []node
	nodeIdx uint16
	search  nodeSearch
	head    *node
	tail    *node
}

func newNodes(config nodesConfig) nodes {
	return nodes{
		config:  config,
		nodes:   make([]node, config.Size),
		nodeIdx: 0,
		search:  make(nodeSearch, config.Size),
		head:    nil,
		tail:    nil,
	}
}

func (n *nodes) unchainNode(nodeRef *node) {
	if nodeRef.previous != nil {
		nodeRef.previous.next = nodeRef.next
	}

	if nodeRef.next != nil {
		nodeRef.next.previous = nodeRef.previous
	}

	if n.head == nodeRef {
		n.head = nodeRef.next
	}

	if n.tail == nodeRef {
		n.tail = nodeRef.previous
	}

	nodeRef.next = nil
	nodeRef.previous = nil
}

func (n *nodes) bumpNode(nodeRef *node) {
	n.unchainNode(nodeRef)

	nodeRef.next = n.head
	nodeRef.lastUse = time.Now()

	if n.head != nil {
		n.head.previous = nodeRef
	}

	if n.tail == nil {
		n.tail = nodeRef
	}

	n.head = nodeRef
}

func (n *nodes) clearNode(nodeRef *node) {
	if nodeRef.name == "" {
		return
	}

	oldConn := nodeRef.conn

	delete(n.search, nodeRef.name)

	n.unchainNode(nodeRef)

	nodeRef.name = ""
	nodeRef.lastUse = time.Time{}
	nodeRef.conn = nil

	// Notify the Conn for deletion
	n.config.OnDelete(oldConn)
}

func (n *nodes) scanExpiredNodes() {
	tail := n.tail
	previous := n.tail
	expireTime := time.Now().Add(-(n.config.Expire))

	for {
		if tail == nil {
			break
		}

		if !tail.lastUse.Before(expireTime) {
			break
		}

		previous = tail.previous

		n.clearNode(tail)

		tail = previous
	}
}

func (n *nodes) setNode(
	name string, nodeRef *node, c udpClient) {
	n.clearNode(nodeRef)

	nodeRef.name = name
	nodeRef.conn = c

	n.bumpNode(nodeRef)

	if n.tail == nil {
		n.tail = nodeRef
	}

	n.search[nodeRef.name] = nodeRef
}

func (n *nodes) Set(name string, c udpClient) error {
	n.scanExpiredNodes()

	_, found := n.search[name]

	if found {
		return ErrNodeAlreadyExisted
	}

	nodeRef := &n.nodes[n.nodeIdx]

	n.setNode(name, nodeRef, c)

	n.nodeIdx = (n.nodeIdx + 1) % n.config.Size

	return nil
}

func (n *nodes) Select(name string) (udpClient, error) {
	node, found := n.search[name]

	if !found {
		return nil, ErrNodeNotFound
	}

	n.bumpNode(node)

	return node.conn, nil
}

func (n *nodes) Expire() {
	n.scanExpiredNodes()
}

func (n *nodes) Clear(name string) error {
	node, found := n.search[name]

	if !found {
		return ErrNodeNotFound
	}

	n.clearNode(node)

	return nil
}

func (n *nodes) ClearAll() {
	tail := n.tail
	pervious := n.tail

	for {
		if tail == nil {
			break
		}

		pervious = tail.previous

		n.clearNode(tail)

		tail = pervious
	}
}
