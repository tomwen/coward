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
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/nickrio/coward/common/locked"
	"github.com/nickrio/coward/roles/common/network/conn"
)

type dispathToMulitClientResult struct {
	rLen  int
	rAddr *net.UDPAddr
	rErr  error
	rData []byte
}

type dummyUDPListenerReadData struct {
	Addr *net.UDPAddr
	Buf  []byte
	Err  error
}

type dummyUDPListenerWriteData struct {
	Data []byte
	Addr net.UDPAddr
}
type dummyUDPListenerWriteResultData struct{}

type dummyUDPListener struct {
	readChan  chan dummyUDPListenerReadData
	writeChan chan dummyUDPListenerWriteData
	closed    locked.Boolean
	writing   []dummyUDPListenerWriteData
}

func (d *dummyUDPListener) WriteToUDP(
	b []byte, addr *net.UDPAddr) (int, error) {
	buf := make([]byte, len(b))

	copy(buf, b)

	d.writing = append(d.writing, dummyUDPListenerWriteData{
		Data: buf,
		Addr: *addr,
	})

	return 0, nil
}

func (d *dummyUDPListener) ReadFromUDP(b []byte) (int, *net.UDPAddr, error) {
	if d.closed.Get() {
		return 0, nil, io.EOF
	}

	data, ok := <-d.readChan

	if !ok {
		d.closed.Set(true)

		return 0, nil, io.EOF
	}

	copy(b, data.Buf)

	return len(data.Buf), data.Addr, data.Err
}

func (d *dummyUDPListener) Close() error {
	if d.closed.Get() {
		return nil
	}

	d.closed.Set(true)

	close(d.readChan)

	return io.EOF
}

type dummyUDPWriteDitcherListener struct {
	readChan  chan dummyUDPListenerReadData
	writeChan chan dummyUDPListenerWriteData
	closed    locked.Boolean
}

func (d *dummyUDPWriteDitcherListener) WriteToUDP(
	b []byte, addr *net.UDPAddr) (int, error) {

	return 0, nil
}

func (d *dummyUDPWriteDitcherListener) ReadFromUDP(
	b []byte) (int, *net.UDPAddr, error) {
	if d.closed.Get() {
		return 0, nil, io.EOF
	}

	data, ok := <-d.readChan

	if !ok {
		d.closed.Set(true)

		return 0, nil, io.EOF
	}

	copy(b, data.Buf)

	return len(data.Buf), data.Addr, data.Err
}

func (d *dummyUDPWriteDitcherListener) Close() error {
	if d.closed.Get() {
		return nil
	}

	d.closed.Set(true)

	close(d.readChan)

	return io.EOF
}

func TestDispatcherDispathOne(t *testing.T) {
	testWait := sync.WaitGroup{}
	sending := []byte("Test data")
	sendingLen := len(sending)
	udpAddr, udpAddrErr := net.ResolveUDPAddr("udp", "127.0.0.2:1000")

	if udpAddrErr != nil {
		t.Error("Failed to generate UDP address due to error:", udpAddrErr)

		return
	}

	l := &dummyUDPListener{
		readChan: make(chan dummyUDPListenerReadData),
		closed:   locked.NewBool(false),
	}

	testWait.Add(1)

	d := New(func(udpConn conn.UDPReadWriteCloser) error {
		defer func() {
			l.Close()

			testWait.Done()
		}()

		buf := make([]byte, 1024)

		rLen, rAddr, rErr := udpConn.ReadFromUDP(buf)

		if rErr != nil {
			t.Error("Failed to read UDP data due to error:", rErr)

			return rErr
		}

		if rLen != sendingLen {
			t.Errorf("Failed to read expected length. Expecting %d, got %d",
				sendingLen, rLen)

			return errors.New("Test failed")
		}

		if !rAddr.IP.Equal(udpAddr.IP) {
			t.Errorf("Failed to read expected Remote Address."+
				"Expecting %s, got %s", udpAddr.IP, rAddr.IP)

			return errors.New("Test failed")
		}

		if rAddr.Port != udpAddr.Port {
			t.Errorf("Failed to read expected Remote Port."+
				"Expecting %d, got %d", udpAddr.Port, rAddr.Port)

			return errors.New("Test failed")
		}

		if rAddr.Zone != udpAddr.Zone {
			t.Errorf("Failed to read expected Remote Zone."+
				"Expecting %s, got %s", udpAddr.Zone, rAddr.Zone)

			return errors.New("Test failed")
		}

		if string(buf[:rLen]) != string(sending) {
			t.Errorf("Failed to read data."+
				"Expecting %d, got %d", sending, buf[:rLen])

			return errors.New("Test failed")
		}

		closeErr := udpConn.Close()

		if closeErr != nil {
			t.Error("Failed to close due to error:", closeErr)

			return closeErr
		}

		rLen, rAddr, rErr = udpConn.ReadFromUDP(buf)

		if rErr != io.EOF {
			t.Error("Failed to expecting EOF error, got:", rErr)

			return errors.New("Test failed")
		}

		if rLen != 0 {
			t.Error("Failed UDP read does not resulting 0 read length, got",
				rLen)

			return errors.New("Test failed")
		}

		if rAddr != nil {
			t.Error("Failed UDP read does not resulting nil address, got",
				rAddr)

			return errors.New("Test failed")
		}

		return nil
	}, 2, 10*time.Second)

	go func() {
		l.readChan <- dummyUDPListenerReadData{
			Addr: udpAddr,
			Buf:  sending,
			Err:  nil,
		}
	}()

	d.Dispatch(l, func(err error) bool {
		return true
	})

	testWait.Wait()
}

func TestDispatcherDispathToMulitClients(t *testing.T) {
	testWait := sync.WaitGroup{}
	udpAddress := map[string][]*dispathToMulitClientResult{
		"127.0.0.2:1000": nil,
		"127.0.0.2:1001": nil,
		"127.0.0.2:1002": nil,
		"127.0.0.2:1003": nil,
	}
	udpAddressKeys := []string{}
	udpAddressLock := sync.Mutex{}
	readed := 0

	l := &dummyUDPListener{
		readChan: make(chan dummyUDPListenerReadData),
		closed:   locked.NewBool(false),
	}

	for key := range udpAddress {
		udpAddressKeys = append(udpAddressKeys, key)
	}

	go func() {
		for _, key := range udpAddressKeys {
			udpAddr, udpAddrErr := net.ResolveUDPAddr("udp", key)

			if udpAddrErr != nil {
				t.Error("Failed to generate UDP address due to error:",
					udpAddrErr)

				return
			}

			l.readChan <- dummyUDPListenerReadData{
				Addr: udpAddr,
				Buf:  append([]byte("Test data"), []byte(key)...),
				Err:  nil,
			}
		}

		for _, key := range udpAddressKeys {
			udpAddr, udpAddrErr := net.ResolveUDPAddr("udp", key)

			if udpAddrErr != nil {
				t.Error("Failed to generate UDP address due to error:",
					udpAddrErr)

				return
			}

			l.readChan <- dummyUDPListenerReadData{
				Addr: udpAddr,
				Buf:  append([]byte("TEST"), []byte(key)...),
				Err:  nil,
			}
		}
	}()

	testWait.Add(len(udpAddress))

	d := New(func(udpConn conn.UDPReadWriteCloser) error {
		defer func() {
			udpAddressLock.Lock()
			defer udpAddressLock.Unlock()

			readed++

			if readed >= len(udpAddress) {
				l.Close()
			}

			testWait.Done()
		}()

		buf := make([]byte, 4096)
		received := 0

		for {
			rLen, rAddr, rErr := udpConn.ReadFromUDP(buf)

			if rErr != nil {
				t.Error("Failed to read UDP data due to error:", rErr)

				return rErr
			}

			err := func() error {
				udpAddressLock.Lock()
				defer udpAddressLock.Unlock()

				rAddString := rAddr.String()

				_, found := udpAddress[rAddString]

				if !found {
					t.Error("Unexpected address:", rAddString)

					return errors.New("Test failed")
				}

				readedData := make([]byte, rLen)

				copy(readedData, buf[:rLen])

				udpAddress[rAddString] = append(
					udpAddress[rAddString],
					&dispathToMulitClientResult{
						rLen:  rLen,
						rAddr: rAddr,
						rErr:  rErr,
						rData: readedData,
					})

				received++

				return nil
			}()

			if err != nil {
				return err
			}

			if received >= 2 {
				break
			}
		}

		return udpConn.Close()
	}, uint16(len(udpAddress)), 10*time.Second)

	d.Dispatch(l, func(arg2 error) bool {
		return true
	})

	testWait.Wait()

	for key, results := range udpAddress {
		if len(results) != 2 {
			t.Errorf("Not all data has been received from %s", key)

			return
		}

		for idx, result := range results {
			if result == nil {
				continue
			}

			if key != result.rAddr.String() {
				t.Errorf("Unexpected Remote Address. Expecting %s, got %s",
					key, result.rAddr)

				return
			}

			switch idx {
			case 0:
				if result.rLen != len("Test data"+key) {
					t.Errorf("Unexpected length. Expecting %d, got %d",
						len("Test data"+key), result.rLen)

					return
				}

				if string(result.rData) != "Test data"+key {
					t.Errorf("Unexpected Data. Expecting %s, got %s",
						"Test data"+key, string(result.rData))

					return
				}

			case 1:
				if result.rLen != len("TEST"+key) {
					t.Errorf("Unexpected length. Expecting %d, got %d",
						len("TEST"+key), result.rLen)

					return
				}

				if string(result.rData) != "TEST"+key {
					t.Errorf("Unexpected Data. Expecting %s, got %s",
						"TEST"+key, string(result.rData))

					return
				}
			}
		}
	}
}

func TestDispatcherDispathAndBumpOld(t *testing.T) {
	testWait := sync.WaitGroup{}
	sendWait := sync.WaitGroup{}
	l := &dummyUDPListener{
		readChan: make(chan dummyUDPListenerReadData),
		closed:   locked.NewBool(false),
	}

	sendWait.Add(1)
	testWait.Add(3)

	go func() {
		defer sendWait.Done()

		udpAddr, _ := net.ResolveUDPAddr("udp", "127.0.0.3:1111")

		l.readChan <- dummyUDPListenerReadData{
			Addr: udpAddr,
			Buf:  []byte(udpAddr.String()),
			Err:  nil,
		}

		udpAddr, _ = net.ResolveUDPAddr("udp", "127.0.0.3:1112")

		l.readChan <- dummyUDPListenerReadData{
			Addr: udpAddr,
			Buf:  []byte(udpAddr.String()),
			Err:  nil,
		}

		l.readChan <- dummyUDPListenerReadData{
			Addr: udpAddr,
			Buf:  []byte(udpAddr.String()),
			Err:  nil,
		}

		udpAddr, _ = net.ResolveUDPAddr("udp", "127.0.0.4:1113")

		l.readChan <- dummyUDPListenerReadData{
			Addr: udpAddr,
			Buf:  []byte(udpAddr.String()),
			Err:  nil,
		}

		l.readChan <- dummyUDPListenerReadData{
			Addr: udpAddr,
			Buf:  []byte(udpAddr.String()),
			Err:  nil,
		}

		testWait.Wait()

		l.Close()
	}()

	d := New(func(udpConn conn.UDPReadWriteCloser) error {
		defer testWait.Done()

		var clientAddr *net.UDPAddr

		readCount := 0
		buf := make([]byte, 4096)

		/*
			How this test case works is, the for loop below will only
			be breaked when we readed two packets from udpConn. And we
			only send one packet to the 127.0.0.3:1111 conn.

			In normal situation, it will stay blocked until another
			packet arrives. However, because of we only allowing 2
			concurrent "connections", the oldest (127.0.0.3:1111) conn
			will be bumpped by new one ("127.0.0.4:1113").

			When that happens, Dispatcher will send a EOF to the bumpped
			conn and break the read loop.
		*/

		for {
			rLen, rAddr, rErr := udpConn.ReadFromUDP(buf)

			if rErr != nil {
				switch readCount {
				case 1:
					if clientAddr.String() != "127.0.0.3:1111" {
						t.Error("UDP client does not read all packets")

						return errors.New("Test failed")
					}

				default:
				}

				if rErr != io.EOF {
					t.Error("Unexpected error: Expecting EOF, got", rErr)
				}

				return rErr
			}

			if rAddr.String() != string(buf[:rLen]) {
				t.Errorf(
					"Failed to receive expected data. Expecting %s, got %s",
					rAddr.String(), string(buf[:rLen]))

				return errors.New("Test failed")
			}

			if clientAddr == nil {
				clientAddr = rAddr
			}

			readCount++

			if readCount >= 2 {
				udpConn.Close()
			}
		}
	}, 2, 10*time.Second)

	d.Dispatch(l, func(arg2 error) bool {
		return true
	})

	sendWait.Wait()
}

func TestDispatcherDispathAndExpire(t *testing.T) {
	testWait := sync.WaitGroup{}
	sendWait := sync.WaitGroup{}
	l := &dummyUDPListener{
		readChan: make(chan dummyUDPListenerReadData),
		closed:   locked.NewBool(false),
	}

	sendWait.Add(1)
	testWait.Add(6)

	go func() {
		defer sendWait.Done()

		udpAddr, _ := net.ResolveUDPAddr("udp", "127.0.0.3:1111")

		l.readChan <- dummyUDPListenerReadData{
			Addr: udpAddr,
			Buf:  []byte(udpAddr.String()),
			Err:  nil,
		}

		time.Sleep(100 * time.Millisecond)

		l.readChan <- dummyUDPListenerReadData{
			Addr: udpAddr,
			Buf:  []byte(udpAddr.String()),
			Err:  nil,
		}

		time.Sleep(100 * time.Millisecond)

		l.readChan <- dummyUDPListenerReadData{
			Addr: udpAddr,
			Buf:  []byte(udpAddr.String()),
			Err:  nil,
		}

		time.Sleep(100 * time.Millisecond)

		l.readChan <- dummyUDPListenerReadData{
			Addr: udpAddr,
			Buf:  []byte(udpAddr.String()),
			Err:  nil,
		}

		time.Sleep(100 * time.Millisecond)

		l.readChan <- dummyUDPListenerReadData{
			Addr: udpAddr,
			Buf:  []byte(udpAddr.String()),
			Err:  nil,
		}

		time.Sleep(100 * time.Millisecond)

		l.readChan <- dummyUDPListenerReadData{
			Addr: udpAddr,
			Buf:  []byte(udpAddr.String()),
			Err:  nil,
		}

		testWait.Wait()

		l.Close()
	}()

	d := New(func(udpConn conn.UDPReadWriteCloser) error {
		defer testWait.Done()

		var clientAddr *net.UDPAddr

		readCount := 0
		buf := make([]byte, 4096)

		for {
			rLen, rAddr, rErr := udpConn.ReadFromUDP(buf)

			if rErr != nil {
				if rErr != io.EOF {
					t.Error("Unexpected error: Expecting EOF, got", rErr)
				}

				return rErr
			}

			if rAddr.String() != string(buf[:rLen]) {
				t.Errorf(
					"Failed to receive expected data. Expecting %s, got %s",
					rAddr.String(), string(buf[:rLen]))

				return errors.New("Test failed")
			}

			if clientAddr == nil {
				clientAddr = rAddr
			}

			readCount++

			if readCount >= 2 {
				udpConn.Close()
			}
		}
	}, 128, 50*time.Millisecond)

	d.Dispatch(l, func(arg2 error) bool {
		return true
	})

	sendWait.Wait()
}

func TestDispatcherDispathAndWrite(t *testing.T) {
	testWait := sync.WaitGroup{}
	sendWait := sync.WaitGroup{}
	l := &dummyUDPListener{
		readChan: make(chan dummyUDPListenerReadData),
		closed:   locked.NewBool(false),
	}

	sendWait.Add(1)

	go func() {
		defer sendWait.Done()

		udpAddr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:1111")

		l.readChan <- dummyUDPListenerReadData{
			Addr: udpAddr,
			Buf:  []byte(udpAddr.String()),
			Err:  nil,
		}

		l.readChan <- dummyUDPListenerReadData{
			Addr: udpAddr,
			Buf:  []byte(udpAddr.String()),
			Err:  nil,
		}

		l.readChan <- dummyUDPListenerReadData{
			Addr: udpAddr,
			Buf:  []byte(udpAddr.String()),
			Err:  nil,
		}

		l.readChan <- dummyUDPListenerReadData{
			Addr: udpAddr,
			Buf:  []byte(udpAddr.String()),
			Err:  nil,
		}

		l.readChan <- dummyUDPListenerReadData{
			Addr: udpAddr,
			Buf:  []byte(udpAddr.String()),
			Err:  nil,
		}

		l.readChan <- dummyUDPListenerReadData{
			Addr: udpAddr,
			Buf:  []byte(udpAddr.String()),
			Err:  nil,
		}

		l.readChan <- dummyUDPListenerReadData{
			Addr: udpAddr,
			Buf:  []byte(udpAddr.String()),
			Err:  nil,
		}

		l.readChan <- dummyUDPListenerReadData{
			Addr: udpAddr,
			Buf:  []byte(udpAddr.String()),
			Err:  nil,
		}

		l.readChan <- dummyUDPListenerReadData{
			Addr: udpAddr,
			Buf:  []byte(udpAddr.String()),
			Err:  nil,
		}

		l.readChan <- dummyUDPListenerReadData{
			Addr: udpAddr,
			Buf:  []byte(udpAddr.String()),
			Err:  nil,
		}

		testWait.Wait()

		l.Close()
	}()

	testWait.Add(1)

	d := New(func(udpConn conn.UDPReadWriteCloser) error {
		defer testWait.Done()

		buf := make([]byte, 4096)
		readLoop := 0

		for {
			rLen, rAddr, rErr := udpConn.ReadFromUDP(buf)

			if rErr != nil {
				if rErr != io.EOF {
					t.Error("Unexpected error: Expecting EOF, got", rErr)
				}

				return rErr
			}

			_, wErr := udpConn.WriteToUDP(buf[:rLen], rAddr)

			if wErr != nil {
				return wErr
			}

			readLoop++

			if readLoop >= 10 {
				break
			}
		}

		udpConn.Close()

		return nil
	}, 1, 10*time.Second)

	d.Dispatch(l, func(arg2 error) bool {
		return true
	})

	for _, wData := range l.writing {
		if wData.Addr.String() != "127.0.0.1:1111" {
			t.Errorf("Unexpected Source Address %s, expecting it to be %s",
				wData.Addr.String(), "127.0.0.1:1111")

			return
		}

		if string(wData.Data) != wData.Addr.String() {
			t.Errorf("Unexpected Data %d, expecting it to be %d",
				wData.Data, []byte(wData.Addr.String()))

			return
		}
	}
}

func TestDispatcherForceClose(t *testing.T) {
	callWait := sync.WaitGroup{}
	l := &dummyUDPListener{
		readChan: make(chan dummyUDPListenerReadData),
		closed:   locked.NewBool(false),
	}

	callWait.Add(1)

	go func() {
		udpAddr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:1111")

		l.readChan <- dummyUDPListenerReadData{
			Addr: udpAddr,
			Buf:  []byte(udpAddr.String()),
			Err:  nil,
		}

		callWait.Wait()

		l.Close()
	}()

	d := New(func(udpConn conn.UDPReadWriteCloser) error {
		buf := make([]byte, 4096)

		defer udpConn.Close()

		// Release wait, to call l.Close()
		callWait.Done()

		for {
			rLen, rAddr, rErr := udpConn.ReadFromUDP(buf)

			if rErr != nil {
				return rErr
			}

			_, wErr := udpConn.WriteToUDP(buf[:rLen], rAddr)

			if wErr != nil {
				return wErr
			}
		}
	}, 1, 10*time.Second)

	d.Dispatch(l, func(arg2 error) bool {
		return true
	})
}

func BenchmarkDispatcherDispatch(b *testing.B) {
	udpAddr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:1111")
	address := []byte(udpAddr.String())
	l := &dummyUDPWriteDitcherListener{
		readChan: make(chan dummyUDPListenerReadData),
	}
	dispatcherCloseWait := sync.WaitGroup{}

	dispatcherCloseWait.Add(1)

	go func() {
		defer dispatcherCloseWait.Done()

		d := New(func(udpConn conn.UDPReadWriteCloser) error {
			buf := make([]byte, 4096)
			loopCount := 0

			defer udpConn.Close()

			for {
				rLen, rAddr, rErr := udpConn.ReadFromUDP(buf)

				if rErr != nil {
					return rErr
				}

				_, wErr := udpConn.WriteToUDP(buf[:rLen], rAddr)

				if wErr != nil {
					return wErr
				}

				loopCount++

				if loopCount >= b.N {
					l.Close()
				}
			}
		}, 1, 10*time.Second)

		d.Dispatch(l, func(arg2 error) bool {
			return true
		})
	}()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.readChan <- dummyUDPListenerReadData{
			Addr: udpAddr,
			Buf:  address,
			Err:  nil,
		}
	}

	dispatcherCloseWait.Wait()
}
