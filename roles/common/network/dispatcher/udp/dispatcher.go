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
	"sync"
	"time"

	"github.com/nickrio/coward/roles/common/network/conn"
)

// Dispatcher is the UDP dispatcher
type Dispatcher interface {
	Dispatch(udpConn conn.UDPReadWriteCloser, quiter func(error) bool) error
}

type dispatcher struct {
	requester     func(conn.UDPReadWriteCloser) error
	maxRequests   uint16
	requestExpire time.Duration
}

// New creates a new UDP Dispatcher
func New(
	requester func(conn.UDPReadWriteCloser) error,
	maxRequest uint16,
	requestExpire time.Duration,
) Dispatcher {
	return &dispatcher{
		requester:     requester,
		maxRequests:   maxRequest,
		requestExpire: requestExpire,
	}
}

func (u *dispatcher) handleSend(
	udpConn conn.UDPReadWriteCloser,
	writeChan chan writeData,
) {
	for {
		wData, chOK := <-writeChan

		if !chOK {
			break
		}

		wLen, wErr := udpConn.WriteToUDP(wData.Data, wData.Addr)

		wData.Result <- writeResult{
			Len: wLen,
			Err: wErr,
		}
	}
}

func (u *dispatcher) handleReceive(
	udpConn conn.UDPReadWriteCloser,
	readedDataChan chan *rawRead,
	readedDataPool chan *rawRead,
	writeChan chan writeData,
	readQueueQuit chan bool,
) error {
	requestWaiter := sync.WaitGroup{}
	nodeRemoveDitchWait := sync.WaitGroup{}
	nodeRemoveChan := make(chan string)
	timeTicker := time.NewTicker(u.requestExpire)
	udpNodes := newNodes(nodesConfig{
		Size:   u.maxRequests,
		Expire: u.requestExpire,
		OnDelete: func(c udpClient) {
			c.Close()
		},
	})

	defer func() {
		timeTicker.Stop()

		nodeRemoveDitchWait.Add(1)

		go func() {
			defer nodeRemoveDitchWait.Done()

			for {
				_, ok := <-nodeRemoveChan

				if !ok {
					break
				}
			}
		}()

		udpNodes.ClearAll()

		requestWaiter.Wait()

		close(nodeRemoveChan)

		nodeRemoveDitchWait.Wait()
	}()

	for {
		select {
		case <-readQueueQuit:
			return nil

		case <-timeTicker.C:
			udpNodes.Expire()

		case removingNode := <-nodeRemoveChan:
			udpNodes.Clear(removingNode)

		case readedData := <-readedDataChan:
			rAddrString := readedData.Addr.String()

			conn, selectErr := udpNodes.Select(rAddrString)

			switch selectErr {
			case nil:

			case ErrNodeNotFound:
				clientConn := newClient(
					readedData.Addr,
					writeChan,
				)

				setErr := udpNodes.Set(rAddrString, clientConn)

				if setErr != nil {
					// Put back 1
					readedDataPool <- readedData

					return setErr
				}

				requestWaiter.Add(1)

				go func(c *client, nodeName string) {
					defer func() {
						c.Delete()

						nodeRemoveChan <- nodeName

						requestWaiter.Done()
					}()

					u.requester(c)
				}(clientConn, rAddrString)

				conn = clientConn

			default:
				// Put back 2
				readedDataPool <- readedData

				return selectErr
			}

			conn.Send(readData{
				Len:  readedData.Len,
				Addr: readedData.Addr,
				Data: readedData.Buf[:],
			})

			// Put back final
			readedDataPool <- readedData
		}
	}
}

func (u *dispatcher) handle(
	udpConn conn.UDPReadWriteCloser, quiter func(error) bool) error {
	readedDataPool := make(chan *rawRead, u.maxRequests)
	readedDataChan := make(chan *rawRead, u.maxRequests)
	writeDataChan := make(chan writeData)
	readQueueQuit := make(chan bool)

	for brdFillIdx := uint16(0); brdFillIdx < u.maxRequests; brdFillIdx++ {
		readedDataPool <- &rawRead{
			Addr: nil,
			Len:  0,
			Buf:  [maxPacketBufferSize]byte{},
		}
	}

	handleWait := sync.WaitGroup{}

	defer handleWait.Wait()

	handleWait.Add(2)

	go func() {
		defer handleWait.Done()

		u.handleSend(udpConn, writeDataChan)
	}()

	go func() {
		defer func() {
			close(readQueueQuit)
			close(writeDataChan)
			close(readedDataChan)
			close(readedDataPool)

			handleWait.Done()
		}()

		u.handleReceive(udpConn, readedDataChan,
			readedDataPool, writeDataChan, readQueueQuit)
	}()

	for {
		newRead := <-readedDataPool

		rLen, rAddr, rErr := udpConn.ReadFromUDP(newRead.Buf[:])

		if rErr != nil {
			readedDataPool <- newRead

			if quiter(rErr) {
				readQueueQuit <- true

				return ErrDispatcherClosed
			}

			continue
		}

		newRead.Addr = rAddr
		newRead.Len = rLen

		// The reader of readedDataChan must put newRead back to
		// readedDataPool or we will quickly run out of pool item
		readedDataChan <- newRead
	}
}

func (u *dispatcher) Dispatch(
	udpConn conn.UDPReadWriteCloser,
	quiter func(error) bool,
) error {
	return u.handle(udpConn, quiter)
}
