package main

import (
	"sync"
)

// concurrencyController controls how many gorutines can run at the same time.
//
// A concurrencyController is created with newConcurrencyController() which specifies the
// maximum number of concurrent operations. Henceforth callers call start() in a separate
// goroutine which will block until the number of concurrent operations is below the
// maximum allowed and will return when the caller can commence their concurrent
// operation. Once the currency operation is complete, the caller calls done() which
// releases up the next waiter sitting on the start() call.
//
// Normally the caller creates many separate goroutines, the number of which may easily
// exceed the concurrency limit, they call start() and are concurrently enabled upto the
// configured maximum. In effect, concurrencyControl uses the goroutine stack as the queue
// of pending goroutines to run.
//
// Idiomatic code looks like this:
//
//	const max = 10
//	const required = 20
//	cc := newConcurrencyController(max)
//	for ix := 0; ix < required; ix++ {
//	    go func() {
//		cc.start()	// Stalls if max exceeded
//		doStuff()
//		cc.done()
//	    }()
//	}
//	cc.wait() // All functions are complete on return
//
// Very limit checking is done to ensure that callers are honoring the correct call
// sequence and avoiding inappropriate calls; e.g.: making more done() calls than start()
// calls is not checked, but it can result in a panic or a deadlock or some other
// undesired behaviour.
//
// The control mechanism is a channel containing the concurrency limit of tokens.  A call
// to start() reads the next token from the channel and a call to done() writes a
// previously read token back to the channel. The caller is responsible for ensuring
// matching calls to start() and done().
type concurrencyController struct {
	sync.RWMutex
	limit   int
	minimum int
	tokens  chan concurrencyToken
}

type concurrencyToken struct{}

// newConcurrencyController creates a concurrencyController which allows a maximum of
// "limit" concurrency goroutines to run at any one time.
func newConcurrencyController(limit int) *concurrencyController {
	cc := &concurrencyController{limit: limit, minimum: limit,
		tokens: make(chan concurrencyToken, limit)}

	for ; limit > 0; limit-- { // With go 1.22, change to "for range limit"
		cc.tokens <- concurrencyToken{}
	}

	return cc
}

// start is normally called by an independent goroutine. start blocks until the number of
// running goroutines is below the limit at which point it returns to the caller. A caller
// calls done() on completion to allow the next goroutine blocked on start() to proceed.
func (cc *concurrencyController) start() {
	cc.Lock()
	defer cc.Unlock()

	<-cc.tokens
	l := len(cc.tokens)
	if l < cc.minimum {
		cc.minimum = l
	}
}

// done informs the concurrencyController that this goroutine has completed and that the
// next goroutine blocked on start() can proceed.
func (cc *concurrencyController) done() {
	cc.tokens <- concurrencyToken{}
}

// limitReached returns true if all tokens have been issued at any time since creation
func (cc *concurrencyController) limitReached() bool {
	cc.RLock()
	defer cc.RUnlock()

	return cc.minimum == 0
}
