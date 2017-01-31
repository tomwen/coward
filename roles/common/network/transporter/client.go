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
	"net"
	"strconv"
	"sync"
	"time"

	ccommon "github.com/nickrio/coward/common"
	"github.com/nickrio/coward/roles/common/network/transporter/common"
	"github.com/nickrio/coward/roles/common/network/transporter/connection"
)

var (
	// ErrClientConnectionWaitTimeout is throwed when client connection
	// request is timed out
	ErrClientConnectionWaitTimeout = common.Error(
		"Connection request is timed out")

	// ErrClientConnectionRequestCanncelled is throwed when client has
	// cancelled the connection request
	ErrClientConnectionRequestCanncelled = common.Error(
		"Connection request has been canncelled")

	// ErrClientDisabled is throwed when current Client is disabled
	ErrClientDisabled = common.Error(
		"Transporter Client is disabled")
)

// Client is the Transporter Client
type Client interface {
	Request(builder HandlerBuilder, option RequestOption) (bool, error)
	Kickoff()
}

// client implements Client
type client struct {
	base

	addr            common.Address
	addrString      string
	disabled        bool
	disabledLock    sync.RWMutex
	connectRetry    uint8
	connectTimeout  time.Duration
	clients         []connection.Client
	idleConnChan    chan connection.Client
	liveConnChan    chan connection.Client
	waitingRequests ccommon.Counter
	averageLoad     ccommon.Averager
	requestWait     sync.WaitGroup
}

// NewClient creates a new Transporter Client
func NewClient(
	host string,
	port uint16,
	idleTimeout time.Duration,
	connectionPersistent bool,
	maxConcurrence uint16,
	connectRetry uint8,
	connectTimeout time.Duration,
	wrapper common.ConnWrapper,
	disrupter common.ConnDisrupter,
) Client {
	c := &client{
		base: base{
			wrapper:              wrapper,
			disrupter:            disrupter,
			connectionPersistent: connectionPersistent,
			idleTimeout:          idleTimeout,
		},
		addr: common.NewAddress(host, port),
		addrString: net.JoinHostPort(host, strconv.FormatUint(
			uint64(port), 10)),
		disabled:        false,
		disabledLock:    sync.RWMutex{},
		connectRetry:    connectRetry,
		connectTimeout:  connectTimeout,
		clients:         make([]connection.Client, maxConcurrence),
		idleConnChan:    make(chan connection.Client, maxConcurrence),
		liveConnChan:    make(chan connection.Client, maxConcurrence),
		waitingRequests: ccommon.NewCounter(0),
		averageLoad:     ccommon.NewLockedAverager(int(maxConcurrence) * 3),
		requestWait:     sync.WaitGroup{},
	}

	for clientID := range c.clients {
		c.clients[clientID] = connection.NewClientConn(
			c.idleTimeout,
			c.connectTimeout,
		)

		c.idleConnChan <- c.clients[clientID]
	}

	return c
}

// setDisable will set the disable status of current client
func (c *client) setDisable(s bool) {
	c.disabledLock.Lock()
	defer c.disabledLock.Unlock()

	c.disabled = s
}

// getDisable will get the disable status of current client
func (c *client) getDisable() bool {
	c.disabledLock.RLock()
	defer c.disabledLock.RUnlock()

	return c.disabled
}

// getConnection gets a free connection from connection pool
func (c *client) getConnection(
	canceller Signal,
	delay func(string, float64, uint64),
) (connection.Client, error) {
	var conn connection.Client
	var cErr error

	// Create a new ticker, and don't use time.Tick because
	// it can cause ticker leak
	ticker := time.NewTicker(c.connectTimeout)
	defer ticker.Stop()

	connectStart := time.Now()

	c.waitingRequests.Add(1)

	defer func() {
		// If connection is failed, then we only update
		// waitingRequests but not averageLoad
		waiting := c.waitingRequests.Remove(1)

		if cErr != nil {
			return
		}

		costedTime := time.Now().Sub(connectStart).Seconds()
		avgDelay := c.averageLoad.AddWithWeight(
			costedTime,
			func(currentAvg float64, size int) int {
				// Make averager flavers better delay
				if currentAvg > costedTime {
					return size / 3
				}

				return 1
			})

		if delay == nil {
			return
		}

		delay(c.addrString, avgDelay, waiting)
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

	for retry := uint8(0); retry < c.connectRetry; retry++ {
		cErr = conn.Connect(c.addr, c.wrapper, c.disrupter)

		if cErr == nil {
			break
		}
	}

	if cErr != nil {
		c.idleConnChan <- conn

		return nil, cErr
	}

	return conn, nil
}

// Request connects the target server and perform query
func (c *client) Request(
	builder HandlerBuilder,
	option RequestOption,
) (bool, error) {
	c.requestWait.Add(1)

	defer c.requestWait.Done()

	var err error

	// Requested indicates if the Request method
	// did send the request to server
	requested := true
	firstTry := true

	for re := uint8(0); re < c.connectRetry; re++ {
		doRetry := false

		requested = true

		err = func() error {
			if c.getDisable() {
				return ErrClientDisabled
			}

			canForceResetTspConn := true
			conn, connErr := c.getConnection(option.Canceller, option.Delay)

			if connErr != nil {
				requested = false

				return connErr
			}

			if firstTry {
				re--
			}

			handler := builder(HandlerConfig{
				Server: conn,
				Buffer: option.Buffer,
			})

			defer func() {
				handler.Close()

				// Close connection AFTER handler closed
				if !c.connectionPersistent {
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

			handlerErr := handler.Handle()

			// If the error is those we care, return those errors
			//
			// Here is some notice on Trying function:
			// It is important to known that we implement handle
			// request Retry by re-running the failed handler.
			// It will lead to some severe problem IF HANDLED
			// INCORRECTLY.
			//
			// The rule is simple though:
			//   We only retry when we 100% sure that user's
			//   request hasn't generated any side effect.
			//
			// For example:
			//   We can retry when:
			//     - User failed to connect to server
			//     - User is using a closed connection
			//   We can't retry when:
			//     - Timeout:   user's request may already been
			//                  proccessed by server
			//     - EOF:       User's request may already finished
			//
			switch handlerErr {
			case nil:
				return nil

			case connection.ErrUnconnectable:
				fallthrough
			case connection.ErrBroken:
				fallthrough
			case connection.ErrNotEstablished:
				doRetry = true
				requested = false

				return handlerErr

			case connection.ErrReadTimeout:
				fallthrough
			case connection.ErrWriteTimeout:
				fallthrough
			case connection.ErrTimeout:
				// Notice here as we treat Timeout error as something
				// serious. The reason behind it is:
				// When a transporter connection is timed out, the
				// progress of that handler may be disrupted. And the
				// server connection which responding to that handler
				// may not be reset and still expecting handler's
				// next command. This will destroy the sync status
				// between server and client.
				//
				// So, to reduce the chance of that problem, we must
				// reset current transporter connection so server will
				// release the session and it's sync status.
				//
				// Why we check this error here instead of letting
				// handler tell us what to do is because another rules
				// states that the Transporter will handle it's own error.
				// Handler had it's own errors to take care of.
				conn.Close()

				canForceResetTspConn = false

				// Also notice here:
				// Even through the Transporter connection will be closed
				// when ErrTimeout happened, the handler can still ask a
				// replay for the request.
				// This is because we don't actually know whether or not
				// to retry for the timeouted request, only hander knows
				// that.
				// If the timeout is happened during Relay stage, target
				// server and source client may already exchanged some
				// data. In this case, retry is dangerous.
				// Otherwise, if the target server and source client hasn't
				// exchanged any data, then it's perfectly safe for us
				// to retry the request when it failed.
				fallthrough

			default:
				wantToRetry, wantToResetTspConn, handledErr := handler.Error(
					handlerErr)

				if handledErr == nil {
					return nil
				}

				retry, resetTspConn, reqErr := option.Error(
					wantToRetry, wantToResetTspConn, handledErr)

				if reqErr == nil {
					return nil
				}

				if retry {
					doRetry = true
				}

				if resetTspConn && canForceResetTspConn {
					conn.Close()
				}

				requested = false

				return reqErr
			}
		}()

		firstTry = false

		if !doRetry {
			break
		}
	}

	return requested, err
}

// Kickoff disconnect active clients from server
func (c *client) Kickoff() {
	c.setDisable(true)
	defer c.setDisable(false)

	breakLoop := false

	for _, client := range c.clients {
		client.Close()
	}

	for {
		if breakLoop {
			break
		}

		select {
		case conn := <-c.liveConnChan:
			c.idleConnChan <- conn

		default:
			breakLoop = true
		}
	}

	c.requestWait.Wait()
}
