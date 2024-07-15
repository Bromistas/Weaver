package chord

import (
	"bytes"
	"log"
	"sort"
)

func (r *Ring) init(conf *Config, trans Transport) {
	// Set our variables
	r.Config = conf
	r.Vnodes = make([]*localVnode, conf.NumVnodes)
	r.Transport = InitLocalTransport(trans)
	r.delegateCh = make(chan func(), 32)

	// Initializes the Vnodes
	for i := 0; i < conf.NumVnodes; i++ {
		vn := &localVnode{}
		r.Vnodes[i] = vn
		vn.Ring = r
		vn.init(i)
	}

	// Sort the Vnodes
	sort.Sort(r)
}

// Len is the number of Vnodes
func (r *Ring) Len() int {
	return len(r.Vnodes)
}

// Less returns whether the vnode with index i should sort
// before the vnode with index j.
func (r *Ring) Less(i, j int) bool {
	return bytes.Compare(r.Vnodes[i].Id, r.Vnodes[j].Id) == -1
}

// Swap swaps the Vnodes with indexes i and j.
func (r *Ring) Swap(i, j int) {
	r.Vnodes[i], r.Vnodes[j] = r.Vnodes[j], r.Vnodes[i]
}

// Returns the nearest local vnode to the key
func (r *Ring) nearestVnode(key []byte) *localVnode {
	for i := len(r.Vnodes) - 1; i >= 0; i-- {
		if bytes.Compare(r.Vnodes[i].Id, key) == -1 {
			return r.Vnodes[i]
		}
	}
	// Return the last vnode
	return r.Vnodes[len(r.Vnodes)-1]
}

// Schedules each vnode in the Ring
func (r *Ring) schedule() {
	if r.Config.Delegate != nil {
		go r.delegateHandler()
	}
	for i := 0; i < len(r.Vnodes); i++ {
		r.Vnodes[i].schedule()
	}
}

// Wait for all the Vnodes to shutdown
func (r *Ring) stopVnodes() {
	r.shutdown = make(chan bool, r.Config.NumVnodes)
	for i := 0; i < r.Config.NumVnodes; i++ {
		<-r.shutdown
	}
}

// Stops the delegate handler
func (r *Ring) stopDelegate() {
	if r.Config.Delegate != nil {
		// Wait for all delegate messages to be processed
		<-r.invokeDelegate(r.Config.Delegate.Shutdown)
		close(r.delegateCh)
	}
}

// Initializes the Vnodes with their local Successors
func (r *Ring) setLocalSuccessors() {
	numV := len(r.Vnodes)
	numSuc := min(r.Config.NumSuccessors, numV-1)
	for idx, vnode := range r.Vnodes {
		for i := 0; i < numSuc; i++ {
			vnode.Successors[i] = &r.Vnodes[(idx+i+1)%numV].Vnode
		}
	}
}

// Invokes a function on the delegate and returns completion channel
func (r *Ring) invokeDelegate(f func()) chan struct{} {
	if r.Config.Delegate == nil {
		return nil
	}

	ch := make(chan struct{}, 1)
	wrapper := func() {
		defer func() {
			ch <- struct{}{}
		}()
		f()
	}

	r.delegateCh <- wrapper
	return ch
}

// This handler runs in a go routine to invoke methods on the delegate
func (r *Ring) delegateHandler() {
	for {
		f, ok := <-r.delegateCh
		if !ok {
			break
		}
		r.safeInvoke(f)
	}
}

// Called to safely call a function on the delegate
func (r *Ring) safeInvoke(f func()) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Caught a panic invoking a delegate function! Got: %s", r)
		}
	}()
	f()
}
