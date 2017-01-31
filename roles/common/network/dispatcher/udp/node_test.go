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
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

type dummyUDPReadWriteCloser struct {
	name string
}

func (d *dummyUDPReadWriteCloser) Send(readData) error {
	return nil
}

func (d *dummyUDPReadWriteCloser) WriteToUDP(
	b []byte, addr *net.UDPAddr) (int, error) {
	return 0, nil
}

func (d *dummyUDPReadWriteCloser) ReadFromUDP(
	b []byte) (int, *net.UDPAddr, error) {
	return 0, nil, nil
}

func (d *dummyUDPReadWriteCloser) Close() error {
	return nil
}

func (d *dummyUDPReadWriteCloser) Delete() {
}

func testPrintFromHeadNodes(node *node) []string {
	nodeNames := []string{}
	head := node

	for {
		if head == nil {
			break
		}

		nodeNames = append(nodeNames, head.name)

		head = head.next
	}

	return nodeNames
}

func testPrintFromTailNodes(node *node) []string {
	nodeNames := []string{}
	tail := node

	for {
		if tail == nil {
			break
		}

		nodeNames = append(nodeNames, tail.name)

		tail = tail.previous
	}

	return nodeNames
}

func testStringArrayEq(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for idx, val := range a {
		if val == b[idx] {
			continue
		}

		return false
	}

	return true
}

func TestNodesQueueSet(t *testing.T) {
	deleted := []string{}
	n := newNodes(nodesConfig{
		Size:   4,
		Expire: 10 * time.Second,
		OnDelete: func(c udpClient) {
			deleted = append(deleted, c.(*dummyUDPReadWriteCloser).name)
		},
	})

	n.Set("#1", &dummyUDPReadWriteCloser{
		name: "#1",
	})

	n.Select("#1")

	if n.tail == nil {
		t.Error("Failed to set tail")

		return
	}

	if n.head == nil {
		t.Error("Failed to set head")

		return
	}

	if n.head != n.tail {
		t.Error("Invalid tail and head")

		return
	}
}

func TestNodesQueueChain(t *testing.T) {
	deleted := []string{}
	n := newNodes(nodesConfig{
		Size:   4,
		Expire: 10 * time.Second,
		OnDelete: func(c udpClient) {
			deleted = append(deleted, c.(*dummyUDPReadWriteCloser).name)
		},
	})

	// Test 1: Set 4 nodes to completely fillup the container,
	// then test the chain orders
	n.Set("#1", &dummyUDPReadWriteCloser{
		name: "#1",
	})

	n.Set("#2", &dummyUDPReadWriteCloser{
		name: "#2",
	})

	n.Set("#3", &dummyUDPReadWriteCloser{
		name: "#3",
	})

	n.Set("#4", &dummyUDPReadWriteCloser{
		name: "#4",
	})

	if !testStringArrayEq(testPrintFromHeadNodes(n.head), []string{
		"#4", "#3", "#2", "#1"}) {
		t.Errorf("Incorrect node chain head order. Expected %s, got %s",
			[]string{"#4", "#3", "#2", "#1"},
			testPrintFromHeadNodes(n.head))

		return
	}

	if !testStringArrayEq(testPrintFromTailNodes(n.tail), []string{
		"#1", "#2", "#3", "#4"}) {
		t.Errorf("Incorrect node chain tail order. Expected %s, got %s",
			[]string{"#1", "#2", "#3", "#4"},
			testPrintFromTailNodes(n.tail))

		return
	}

	// Test 2: try to use #2 object in chain. It should be move to the
	// chain head
	n.Select("#2")

	if !testStringArrayEq(testPrintFromHeadNodes(n.head), []string{
		"#2", "#4", "#3", "#1"}) {
		t.Errorf("Incorrect node chain head order. Expected %s, got %s",
			[]string{"#2", "#4", "#3", "#1"},
			testPrintFromHeadNodes(n.head))

		return
	}

	if !testStringArrayEq(testPrintFromTailNodes(n.tail), []string{
		"#1", "#3", "#4", "#2"}) {
		t.Errorf("Incorrect node chain tail order. Expected %s, got %s",
			[]string{"#1", "#3", "#4", "#2"},
			testPrintFromTailNodes(n.tail))

		return
	}

	// Test 3: Add other one, #1 node should be popped because we out of
	// container space
	n.Set("#5", &dummyUDPReadWriteCloser{
		name: "#5",
	})

	if !testStringArrayEq(testPrintFromHeadNodes(n.head), []string{
		"#5", "#2", "#4", "#3"}) {
		t.Errorf("Incorrect node chain head order. Expected %s, got %s",
			[]string{"#5", "#2", "#4", "#3"},
			testPrintFromHeadNodes(n.head))

		return
	}

	if !testStringArrayEq(testPrintFromTailNodes(n.tail), []string{
		"#3", "#4", "#2", "#5"}) {
		t.Errorf("Incorrect node chain tail order. Expected %s, got %s",
			[]string{"#3", "#4", "#2", "#5"},
			testPrintFromTailNodes(n.tail))

		return
	}

	if len(deleted) != 1 {
		t.Errorf("Expected 1 conn will be deleted, got %d", len(deleted))

		return
	}

	if deleted[0] != "#1" {
		t.Errorf("Expected #1 will be deleted, got %s", deleted[0])

		return
	}

	deleted = []string{}

	// Test 4: Delete tail (#3)
	n.Clear("#3")

	if !testStringArrayEq(testPrintFromHeadNodes(n.head), []string{
		"#5", "#2", "#4"}) {
		t.Errorf("Incorrect node chain head order. Expected %s, got %s",
			[]string{"#5", "#2", "#4"},
			testPrintFromHeadNodes(n.head))

		return
	}

	if !testStringArrayEq(testPrintFromTailNodes(n.tail), []string{
		"#4", "#2", "#5"}) {
		t.Errorf("Incorrect node chain tail order. Expected %s, got %s",
			[]string{"#4", "#2", "#5"},
			testPrintFromTailNodes(n.tail))

		return
	}

	if len(deleted) != 1 {
		t.Errorf("Expected 1 conn will be deleted, got %d", len(deleted))

		return
	}

	if deleted[0] != "#3" {
		t.Errorf("Expected #3 will be deleted, got %s", deleted[0])

		return
	}

	deleted = []string{}

	// Test 5: Delete head
	n.Clear("#5")

	if !testStringArrayEq(testPrintFromHeadNodes(n.head), []string{
		"#2", "#4"}) {
		t.Errorf("Incorrect node chain head order. Expected %s, got %s",
			[]string{"#2", "#4"},
			testPrintFromHeadNodes(n.head))

		return
	}

	if !testStringArrayEq(testPrintFromTailNodes(n.tail), []string{
		"#4", "#2"}) {
		t.Errorf("Incorrect node chain tail order. Expected %s, got %s",
			[]string{"#4", "#2"},
			testPrintFromTailNodes(n.tail))

		return
	}

	if len(deleted) != 1 {
		t.Errorf("Expected 1 conn will be deleted, got %d", len(deleted))

		return
	}

	if deleted[0] != "#5" {
		t.Errorf("Expected #5 will be deleted, got %s", deleted[0])

		return
	}

	deleted = []string{}

	// Test 6: Add a new node so we have 3 nodes, then delete the middle one
	n.Set("#7", &dummyUDPReadWriteCloser{
		name: "#7",
	})

	n.Clear("#2")

	if !testStringArrayEq(testPrintFromHeadNodes(n.head), []string{
		"#7", "#4"}) {
		t.Errorf("Incorrect node chain head order. Expected %s, got %s",
			[]string{"#7", "#4"},
			testPrintFromHeadNodes(n.head))

		return
	}

	if !testStringArrayEq(testPrintFromTailNodes(n.tail), []string{
		"#4", "#7"}) {
		t.Errorf("Incorrect node chain tail order. Expected %s, got %s",
			[]string{"#4", "#7"},
			testPrintFromTailNodes(n.tail))

		return
	}

	if len(deleted) != 1 {
		t.Errorf("Expected 1 conn will be deleted, got %d", len(deleted))

		return
	}

	if deleted[0] != "#2" {
		t.Errorf("Expected #2 will be deleted, got %s", deleted[0])

		return
	}

	deleted = []string{}

	// Test 7: Add another three, see delete order of the old nodes
	n.Set("#8", &dummyUDPReadWriteCloser{
		name: "#8",
	})

	n.Set("#9", &dummyUDPReadWriteCloser{
		name: "#9",
	})

	n.Set("#10", &dummyUDPReadWriteCloser{
		name: "#10",
	})

	n.Set("#11", &dummyUDPReadWriteCloser{
		name: "#11",
	})

	if !testStringArrayEq(testPrintFromHeadNodes(n.head), []string{
		"#11", "#10", "#9", "#8"}) {
		t.Errorf("Incorrect node chain head order. Expected %s, got %s",
			[]string{"#11", "#10", "#9", "#8"},
			testPrintFromHeadNodes(n.head))

		return
	}

	if !testStringArrayEq(testPrintFromTailNodes(n.tail), []string{
		"#8", "#9", "#10", "#11"}) {
		t.Errorf("Incorrect node chain tail order. Expected %s, got %s",
			[]string{"#8", "#9", "#10", "#11"},
			testPrintFromTailNodes(n.tail))

		return
	}

	if len(deleted) != 2 {
		t.Errorf("Expected 2 conn will be deleted, got %d", len(deleted))

		return
	}

	if strings.Join(deleted, ",") != "#4,#7" {
		t.Errorf("Expected #4, #7 will be deleted, got %s", deleted)

		return
	}
}

func TestNodesExpire(t *testing.T) {
	deleted := []string{}
	n := newNodes(nodesConfig{
		Size:   4,
		Expire: 10 * time.Millisecond,
		OnDelete: func(c udpClient) {
			deleted = append(deleted, c.(*dummyUDPReadWriteCloser).name)
		},
	})
	loopLimit := 0
	lastSetName := "Test"

	n.Set("Test", &dummyUDPReadWriteCloser{
		name: "Test",
	})

	for {
		if loopLimit > 10 {
			break
		}

		loopLimit++

		time.Sleep(11 * time.Millisecond)

		setErr := n.Set(
			fmt.Sprintf("Test %d", loopLimit), &dummyUDPReadWriteCloser{
				name: fmt.Sprintf("Test %d", loopLimit),
			})

		if setErr != nil {
			t.Errorf("Unexpected Set error: %s", setErr)

			return
		}

		if len(deleted) != 1 {
			t.Errorf("Expecting 1 node will be deleted, got %d (%s)",
				len(deleted), deleted)

			return
		}

		if deleted[0] != lastSetName {
			t.Errorf("Expecting deleted node is %s, got %s",
				lastSetName, deleted[0])

			return
		}

		deleted = []string{}

		lastSetName = fmt.Sprintf("Test %d", loopLimit)
	}
}

func TestNodesSelectUndefined(t *testing.T) {
	n := newNodes(nodesConfig{
		Size:     4,
		Expire:   10 * time.Millisecond,
		OnDelete: func(c udpClient) {},
	})

	conn, selectErr := n.Select("#UNDEFINED")

	if selectErr != ErrNodeNotFound {
		t.Errorf("Failed to get expected error: ErrNodeNotFound, got %s",
			selectErr)

		return
	}

	if conn != nil {
		t.Error("Failed to expect a nil result from failed Select")

		return
	}
}

func TestNodesDoubleSet(t *testing.T) {
	n := newNodes(nodesConfig{
		Size:     4,
		Expire:   10 * time.Millisecond,
		OnDelete: func(c udpClient) {},
	})

	setErr := n.Set("#TEST", &dummyUDPReadWriteCloser{})

	if setErr != nil {
		t.Errorf("Failed when trying to set Node: %s", setErr)

		return
	}

	setErr = n.Set("#TEST", &dummyUDPReadWriteCloser{})

	if setErr != ErrNodeAlreadyExisted {
		t.Errorf("Expecting error ErrNodeAlreadyExisted, got: %s",
			setErr)

		return
	}
}

func TestNodesClear(t *testing.T) {
	deleted := []string{}
	n := newNodes(nodesConfig{
		Size:   4,
		Expire: 100 * time.Millisecond,
		OnDelete: func(c udpClient) {
			deleted = append(deleted, c.(*dummyUDPReadWriteCloser).name)
		},
	})

	n.Set("#1", &dummyUDPReadWriteCloser{})
	n.Set("#2", &dummyUDPReadWriteCloser{})
	n.Set("#3", &dummyUDPReadWriteCloser{})
	n.Set("#4", &dummyUDPReadWriteCloser{})

	n.ClearAll()

	if !testStringArrayEq(testPrintFromHeadNodes(n.head), []string{}) {
		t.Errorf("Incorrect node chain head order. Expected %s, got %s",
			[]string{},
			testPrintFromHeadNodes(n.head))

		return
	}
}

func BenchmarkNodeSelect(b *testing.B) {
	n := newNodes(nodesConfig{
		Size:     4,
		Expire:   100 * time.Millisecond,
		OnDelete: func(c udpClient) {},
	})
	names := []string{
		"#1", "#2", "#3", "#4",
	}
	selectNameIdx := 0

	for _, v := range names {
		n.Set(v, &dummyUDPReadWriteCloser{})
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, e := n.Select(names[selectNameIdx]); e != nil {
			b.Error("Error:", e)
		}

		selectNameIdx = (selectNameIdx + 1) % 4
	}
}

func BenchmarkNodeSet(b *testing.B) {
	n := newNodes(nodesConfig{
		Size:     4,
		Expire:   100 * time.Millisecond,
		OnDelete: func(c udpClient) {},
	})
	names := []string{
		"#1", "#2", "#3", "#4", "#5",
	}
	dummy := &dummyUDPReadWriteCloser{}
	setNameIdx := 0

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if e := n.Set(names[setNameIdx], dummy); e != nil {
			b.Error("Error:", e)
		}

		setNameIdx = (setNameIdx + 1) % 5
	}
}

func BenchmarkNodeSetSelect(b *testing.B) {
	n := newNodes(nodesConfig{
		Size:     4,
		Expire:   100 * time.Millisecond,
		OnDelete: func(c udpClient) {},
	})
	names := []string{
		"#1", "#2", "#3", "#4",
	}
	dummy := &dummyUDPReadWriteCloser{}
	selectNameIdx := 0
	setNameIdx := 0

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		n.Select(names[selectNameIdx])
		n.Set(names[setNameIdx], dummy)

		selectNameIdx = (selectNameIdx + 1) % 4
		setNameIdx = (setNameIdx + selectNameIdx) % 4
	}
}
