package chord

import (
	"bytes"
	"crypto/sha1"
	"sort"
	"testing"
	"time"
)

type MockDelegate struct {
	shutdown bool
}

func (m *MockDelegate) NewPredecessor(local, remoteNew, remotePrev *Vnode) {
}
func (m *MockDelegate) Leaving(local, pred, succ *Vnode) {
}
func (m *MockDelegate) PredecessorLeaving(local, remote *Vnode) {
}
func (m *MockDelegate) SuccessorLeaving(local, remote *Vnode) {
}
func (m *MockDelegate) Shutdown() {
	m.shutdown = true
}

func makeRing() *Ring {
	conf := &Config{
		NumVnodes:     5,
		NumSuccessors: 8,
		HashFunc:      sha1.New,
		Hashbits:      160,
		StabilizeMin:  time.Second,
		StabilizeMax:  5 * time.Second,
	}

	ring := &Ring{}
	ring.init(conf, nil)
	return ring
}

func TestRingInit(t *testing.T) {
	// Create a Ring
	ring := &Ring{}
	conf := DefaultConfig("test")
	ring.init(conf, nil)

	// Test features
	if ring.Config != conf {
		t.Fatalf("wrong Config")
	}
	if ring.Transport == nil {
		t.Fatalf("missing Transport")
	}

	// Check the Vnodes
	for i := 0; i < conf.NumVnodes; i++ {
		if ring.Vnodes[i] == nil {
			t.Fatalf("missing vnode!")
		}
		if ring.Vnodes[i].Ring != ring {
			t.Fatalf("Ring missing!")
		}
		if ring.Vnodes[i].Id == nil {
			t.Fatalf("ID not initialized!")
		}
	}
}

func TestRingLen(t *testing.T) {
	ring := makeRing()
	if ring.Len() != 5 {
		t.Fatalf("wrong len")
	}
}

func TestRingSort(t *testing.T) {
	ring := makeRing()
	sort.Sort(ring)
	if bytes.Compare(ring.Vnodes[0].Id, ring.Vnodes[1].Id) != -1 {
		t.Fatalf("bad sort")
	}
	if bytes.Compare(ring.Vnodes[1].Id, ring.Vnodes[2].Id) != -1 {
		t.Fatalf("bad sort")
	}
	if bytes.Compare(ring.Vnodes[2].Id, ring.Vnodes[3].Id) != -1 {
		t.Fatalf("bad sort")
	}
	if bytes.Compare(ring.Vnodes[3].Id, ring.Vnodes[4].Id) != -1 {
		t.Fatalf("bad sort")
	}
}

func TestRingNearest(t *testing.T) {
	ring := makeRing()
	ring.Vnodes[0].Id = []byte{2}
	ring.Vnodes[1].Id = []byte{4}
	ring.Vnodes[2].Id = []byte{7}
	ring.Vnodes[3].Id = []byte{10}
	ring.Vnodes[4].Id = []byte{14}
	key := []byte{6}

	near := ring.nearestVnode(key)
	if near != ring.Vnodes[1] {
		t.Fatalf("got wrong node back!")
	}

	key = []byte{0}
	near = ring.nearestVnode(key)
	if near != ring.Vnodes[4] {
		t.Fatalf("got wrong node back!")
	}
}

func TestRingSchedule(t *testing.T) {
	ring := makeRing()
	ring.setLocalSuccessors()
	ring.schedule()
	for i := 0; i < len(ring.Vnodes); i++ {
		if ring.Vnodes[i].timer == nil {
			t.Fatalf("expected timer!")
		}
	}
	ring.stopVnodes()
}

func TestRingSetLocalSucc(t *testing.T) {
	ring := makeRing()
	ring.setLocalSuccessors()
	for i := 0; i < len(ring.Vnodes); i++ {
		for j := 0; j < 4; j++ {
			if ring.Vnodes[i].Successors[j] == nil {
				t.Fatalf("expected successor!")
			}
		}
		if ring.Vnodes[i].Successors[4] != nil {
			t.Fatalf("should not have 5th successor!")
		}
	}

	// Verify the successor manually for node 3
	vn := ring.Vnodes[2]
	if vn.Successors[0] != &ring.Vnodes[3].Vnode {
		t.Fatalf("bad succ!")
	}
	if vn.Successors[1] != &ring.Vnodes[4].Vnode {
		t.Fatalf("bad succ!")
	}
	if vn.Successors[2] != &ring.Vnodes[0].Vnode {
		t.Fatalf("bad succ!")
	}
	if vn.Successors[3] != &ring.Vnodes[1].Vnode {
		t.Fatalf("bad succ!")
	}
}

func TestRingDelegate(t *testing.T) {
	d := &MockDelegate{}
	ring := makeRing()
	ring.setLocalSuccessors()
	ring.Config.Delegate = d
	ring.schedule()

	var b bool
	f := func() {
		println("run!")
		b = true
	}
	ch := ring.invokeDelegate(f)
	if ch == nil {
		t.Fatalf("expected chan")
	}
	select {
	case <-ch:
	case <-time.After(time.Second):
		t.Fatalf("timeout")
	}
	if !b {
		t.Fatalf("b should be true")
	}

	ring.stopDelegate()
	if !d.shutdown {
		t.Fatalf("delegate did not get shutdown")
	}
}
