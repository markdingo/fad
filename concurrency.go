package main

import (
	"sync"
)

// concurrencyControl controls how many concurrent operations of some type can occur at
// the same time. A concurrencyController is created with newConcurrencyController() which
// specifies the maximum number of concurrent operations. Henceforth callers call start()
// which will block until the number of concurrent operations is below the maximum allowed
// and will return when the caller can commence their concurrent operation. Once the
// currency operation is complete, the caller calls done() which releases up the next
// waiter sitting on the start() call.
//
// Normally the caller creates many separate goroutines, the number of which may easily
// exceed the concurrency limit, they call start() and are concurrently enabled upto the
// configured maximum.
//
// Very limit checking is done to ensure that callers are honoring the correct call
// sequence and inappropriate calls, such as more done() calls than start() calls can
// result in a panic or a deadlock or some other undesired behaviour.
//
// The control mechanism is a channel containing the concurrency limit of tokens.  A call
// to start() reads the next token from the channel and a call to done() writes a
// previously read token back to the channel. The caller is responsible for ensuring
// matching calls to start() and done().

type concurrencyToken struct{}

type concurrencyController struct {
	sync.RWMutex
	limit   int
	minimum int
	tokens  chan concurrencyToken
}

func newConcurrencyController(limit int) *concurrencyController {
	cc := &concurrencyController{limit: limit, minimum: limit,
		tokens: make(chan concurrencyToken, limit)}

	for ; limit > 0; limit-- { // With go 1.22, change to "for range limit"
		cc.tokens <- concurrencyToken{}
	}

	return cc
}

func (cc *concurrencyController) start() {
	cc.Lock()
	defer cc.Unlock()

	<-cc.tokens
	l := len(cc.tokens)
	if l < cc.minimum {
		cc.minimum = l
	}
}

func (cc *concurrencyController) done() {
	cc.tokens <- concurrencyToken{}
}

// Return true if all tokens have been issued at any time since creation
func (cc *concurrencyController) limitReached() bool {
	cc.RLock()
	defer cc.RUnlock()

	return cc.minimum == 0
}
