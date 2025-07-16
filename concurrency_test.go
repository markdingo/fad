package main

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestConcurrency(t *testing.T) {
	const (
		limit          = 5
		multiplier     = 4
		sleepFor       = time.Second / 10 // 1/10th of a second should be plenty
		expectedElapse = 2 * sleepFor * multiplier
	)

	var (
		concurrent int32
		wg         sync.WaitGroup
	)

	cc := newConcurrencyController(limit)
	startTime := time.Now()

	if len(cc.tokens) != limit {
		t.Error("New controller has wrong limit of", len(cc.tokens), "expected", limit)
	}
	if cap(cc.tokens) != limit {
		t.Error("New controller has wrong capacity of", cap(cc.tokens), "expected", limit)
	}

	for ix := 0; ix < limit*multiplier; ix++ { // go 1.22 for ix := range limit*multiplier
		wg.Add(1)
		go func(myx int) {
			cc.start()
			time.Sleep(sleepFor)
			max := atomic.AddInt32(&concurrent, 1)
			//			fmt.Println(myx, "Active for", max, "after", time.Now().Sub(startTime).Seconds())
			if max > limit {
				t.Error("Max Inc", max, "GT limit", limit)
			}
			time.Sleep(sleepFor)
			max = atomic.AddInt32(&concurrent, -1)
			if max > limit {
				t.Error("Max Dec", max, "GT limit", limit)
			}
			cc.done()
			wg.Done()
		}(ix)
	}
	time.Sleep(sleepFor) // All tokens should be consumed after this
	l := len(cc.tokens)
	if l != 0 {
		t.Error(l, "tokens available, expected zero")
	}
	wg.Wait()

	if len(cc.tokens) != cap(cc.tokens) {
		t.Error("All tokens should have been returned", len(cc.tokens), cap(cc.tokens))
	}

	// Should have waited for at least expectedElapse time

	dur := time.Now().Sub(startTime)
	if dur < expectedElapse {
		t.Error("Expected to wait for at least", expectedElapse, "not", dur)
	}

	// Should have also exhausted all tokens at some point

	if !cc.limitReached() {
		t.Error("Expected to consume all tokens at some stage during test")
	}
}
