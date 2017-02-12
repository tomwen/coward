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

package transporter

import (
	"errors"
	"sync"
	"time"

	ccommon "github.com/nickrio/coward/common"
	"github.com/nickrio/coward/common/locked"
)

// Transporter client general errors
var (
	ErrClientConnectionWaitTimeout = errors.New(
		"Transporter Connection waiting is timed out")

	ErrClientConnectionRequestCanncelled = errors.New(
		"Transporter Connection request has been canncelled")

	ErrClientDisabled = errors.New(
		"Transporter Client is disabled")
)

// Transporter client status
var (
	ErrClientInitialConnectionFailed = NewError(
		"Transporter Client failed to establish initial connection with server")
)

// Client is the Transporter Client
type Client interface {
	Request(builder HandlerBuilder, option RequestOption) (bool, error)
	Kickoff()
}

// client implements Client
type client struct {
	retry           uint8
	waitTimeout     time.Duration
	disabled        locked.Boolean
	reuseConn       bool
	clients         []ClientConn
	idleConnChan    chan ClientConn
	liveConnChan    chan ClientConn
	waitingRequests ccommon.Counter
	avgConnSelDelay ccommon.Averager
	requestWait     sync.WaitGroup
}

// NewClient creates a new Transporter client
func NewClient(
	clientBuilder ClientConnBuilder,
	waitTimeout time.Duration,
	concurrence uint16,
	retry uint8,
	reuseConn bool,
) Client {
	c := &client{
		retry:           retry,
		waitTimeout:     waitTimeout,
		disabled:        locked.NewBool(false),
		reuseConn:       reuseConn,
		clients:         make([]ClientConn, concurrence),
		idleConnChan:    make(chan ClientConn, concurrence),
		liveConnChan:    make(chan ClientConn, concurrence),
		waitingRequests: ccommon.NewCounter(0),
		avgConnSelDelay: ccommon.NewLockedAverager(int(concurrence)),
		requestWait:     sync.WaitGroup{},
	}

	for clientID := range c.clients {
		c.clients[clientID] = clientBuilder()

		c.idleConnChan <- c.clients[clientID]
	}

	return c
}

// getConnection gets a free connection from connection pool
func (c *client) getConnection(
	canceller Signal,
	delay func(float64, uint64),
) (ClientConn, error) {
	var conn ClientConn
	var cErr error

	// Create a new ticker, and don't use time.Tick because
	// it can cause ticker leak
	ticker := time.NewTicker(c.waitTimeout * time.Duration(c.retry))
	defer ticker.Stop()

	connectStart := time.Now()

	c.waitingRequests.Add(1)

	defer func() {
		// If connection is failed, then we only update
		// waitingRequests but not avgConnSelDelay
		waiting := c.waitingRequests.Remove(1)

		if cErr != nil {
			return
		}

		costedTime := time.Now().Sub(connectStart).Seconds()
		avgDelay := c.avgConnSelDelay.AddWithWeight(
			costedTime,
			func(currentAvg float64, size int) int {
				weight := 1

				// Make averager flavers better delay
				if currentAvg > costedTime {
					weight = size / 5
				}

				if weight < 1 {
					weight = 1
				}

				return weight
			})

		if delay == nil {
			return
		}

		// Feedback connection select delay information with
		// a nice callback
		delay(avgDelay, waiting)
	}()

	select {
	case conn = <-c.liveConnChan:
		// Do nothing

	case conn = <-c.idleConnChan:
		// Do nothing

	case <-ticker.C:
		return nil, ErrClientConnectionWaitTimeout

	case cErr = <-canceller:
		if cErr == nil {
			return nil, ErrClientConnectionRequestCanncelled
		}

		return nil, cErr
	}

	conn.Rewind()

	if conn.Connected() {
		return conn, nil
	}

	// Try at least once
	retry := c.retry

	for {
		cErr = conn.Dial()

		if cErr == nil {
			break
		}

		if retry > 1 {
			retry--

			continue
		}

		break
	}

	if cErr != nil {
		c.idleConnChan <- conn

		return nil, cErr
	}

	return conn, nil
}

// request fetchs a available connection, and send request with it
// returns NeedsRetry bool, error error
func (c *client) request(
	builder HandlerBuilder,
	opt RequestOption,
) (bool, error) {
	var handler Handler

	forceDisconnect := false

	conn, connErr := c.getConnection(opt.Canceller, opt.Delay)

	if connErr != nil {
		return false, UnderError(ErrClientInitialConnectionFailed, connErr)
	}

	defer func() {
		// Making sure connections are closed AFTER Request Handler
		// closed
		if handler != nil {
			handler.Close()
		}

		if forceDisconnect {
			conn.Close()
		} else if !c.reuseConn {
			c.waitingRequests.Load(func(waitingRequest uint64) {
				if waitingRequest > 0 {
					return
				}

				conn.Close()
			})
		}

		if conn.Connected() {
			c.liveConnChan <- conn
		} else {
			c.idleConnChan <- conn
		}
	}()

	handler = builder(HandlerConfig{
		Server: &wrapped{ReadWriteCloser: conn},
		Buffer: opt.Buffer,
	})

	handlerErr := handler.Handle()

	// Rule:
	//
	// 1, If it's a transporter error, then the transporter connection
	//    must be reset (Disconnected)
	// 2, Whether or not to retry is depends on handler.Error and
	//    opt.Error
	//
	_, isTSPErr := handlerErr.(Error)

	if isTSPErr {
		forceDisconnect = true
	}

	wantToRetry, wantToResetTspConn, handledErr := handler.Error(handlerErr)

	if wantToResetTspConn {
		forceDisconnect = true
	}

	if handledErr == nil {
		return false, nil
	}

	retry, resetTspConn, reqErr := opt.Error(
		wantToRetry, wantToResetTspConn, handledErr)

	if resetTspConn {
		forceDisconnect = true
	}

	if reqErr == nil {
		return false, nil
	}

	return retry, reqErr
}

// Request connects the target server and perform query
func (c *client) Request(
	builder HandlerBuilder, option RequestOption) (bool, error) {
	var err error

	if c.disabled.Get() {
		return false, ErrClientDisabled
	}

	c.requestWait.Add(1)
	defer c.requestWait.Done()

	retry := c.retry
	needRetry := false

	for {
		needRetry, err = c.request(builder, option)

		if needRetry && retry > 1 {
			retry--

			continue
		}

		break
	}

	// Treat !needRetry (No need to retry) as requested
	return !needRetry, err
}

// Kickoff disconnect active clients from server
func (c *client) Kickoff() {
	c.disabled.Set(true)
	defer c.disabled.Set(false)

	breakLoop := false

	for _, client := range c.clients {
		client.Close()
	}

	for !breakLoop {
		select {
		case conn := <-c.liveConnChan:
			c.idleConnChan <- conn

		default:
			breakLoop = true
		}
	}

	c.requestWait.Wait()
}
