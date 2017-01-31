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

package network

import (
	"net"
	"sync"
	"testing"
	"time"
)

type dummyConn struct{}

func (d *dummyConn) Read(b []byte) (n int, err error) {
	return 0, nil
}

func (d *dummyConn) Write(b []byte) (n int, err error) {
	return 0, nil
}

func (d *dummyConn) Close() error {
	return nil
}

func (d *dummyConn) LocalAddr() net.Addr {
	return nil
}

func (d *dummyConn) RemoteAddr() net.Addr {
	return nil
}

func (d *dummyConn) SetDeadline(t time.Time) error {
	return nil
}

func (d *dummyConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (d *dummyConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func TestConnectionsPutGetDel(t *testing.T) {
	csWait := sync.WaitGroup{}
	cs := NewConnections()

	csWait.Add(1)

	go func() {
		defer csWait.Done()

		cs.Serve()
	}()

	defer func() {
		cs.Close()

		csWait.Wait()
	}()

	cs.Put("C1", &dummyConn{})
	cs.Put("C2", &dummyConn{})

	conn, getErr := cs.Get("C NOTFOUND")

	if getErr != ErrConnectionsConnectionNotFound {
		t.Error("Expecting error to be ErrConnectionsConnectionNotFound, got",
			getErr)

		return
	}

	conn, getErr = cs.Get("C1")

	if getErr != nil {
		t.Error("Failed to get Connections item due to error",
			getErr)

		return
	}

	if conn == nil {
		t.Error("Failed to get Connections item")

		return
	}

	cs.Del("C1")

	conn, getErr = cs.Get("C1")

	if getErr != ErrConnectionsConnectionNotFound {
		t.Error("Expecting error to be ErrConnectionsConnectionNotFound, got",
			getErr)

		return
	}

	conn, getErr = cs.Get("C2")

	if getErr != nil {
		t.Error("Failed to get Connections item due to error",
			getErr)

		return
	}

	if conn == nil {
		t.Error("Failed to get Connections item")

		return
	}

	// Try again, get request should not create anything
	conn, getErr = cs.Get("C2")

	if getErr != nil {
		t.Error("Failed to get Connections item due to error",
			getErr)

		return
	}

	if conn == nil {
		t.Error("Failed to get Connections item")

		return
	}

	putErr := cs.Put("C2", &dummyConn{})

	if putErr != ErrConnectionsConnectionAlreadyExisted {
		t.Error("Failed to expecting error "+
			"ErrConnectionsConnectionAlreadyExisted, got", putErr)

		return
	}

	_, delErr := cs.Del("C NOT EXISTED")

	if delErr != ErrConnectionsConnectionNotFound {
		t.Error("Failed to expecting error "+
			"ErrConnectionsConnectionNotFound, got", putErr)

		return
	}
}

func TestConnectionsIterate(t *testing.T) {
	csWait := sync.WaitGroup{}
	cs := NewConnections()

	csWait.Add(1)

	go func() {
		defer csWait.Done()

		cs.Serve()
	}()

	defer func() {
		cs.Close()

		csWait.Wait()
	}()

	cs.Put("C1", &dummyConn{})

	csItems := map[string]net.Conn{}

	cs.Iterate(func(name string, conn net.Conn) {
		csItems[name] = conn
	})

	if len(csItems) != 1 {
		t.Error("Invalid amount of Connections items, "+
			"expecting to be %d, got %d", 1, len(csItems))

		return
	}

	_, found := csItems["C1"]

	if !found {
		t.Error("Failed to iterate all connection items")

		return
	}

	cs.Put("C2", &dummyConn{})

	cs.Iterate(func(name string, conn net.Conn) {
		csItems[name] = conn
	})

	if len(csItems) != 2 {
		t.Error("Invalid amount of Connections items, "+
			"expecting to be %d, got %d", 2, len(csItems))

		return
	}

	_, foundC1 := csItems["C1"]
	_, foundC2 := csItems["C2"]

	if !foundC1 || !foundC2 {
		t.Error("Failed to iterate all connection items")

		return
	}
}

func BenchmarkConnections(b *testing.B) {
	csWait := sync.WaitGroup{}
	cs := NewConnections()

	csWait.Add(1)

	go func() {
		defer csWait.Done()

		cs.Serve()
	}()

	defer func() {
		cs.Close()

		csWait.Wait()
	}()

	d := &dummyConn{}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cs.Put("TEST", d)
		cs.Get("TEST")
		cs.Del("TEST")
	}
}
