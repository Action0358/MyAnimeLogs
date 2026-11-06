package common

import (
	"bytes"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRecoverGoroutine_RecoversPanicAndLogs(t *testing.T) {
	var buf bytes.Buffer
	logger := setupTestLogger(&buf)

	done := make(chan struct{})
	go func() {
		defer close(done)
		defer RecoverGoroutine(logger, "test worker")
		panic("boom")
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("goroutine did not finish")
	}

	out := buf.String()
	assert.Contains(t, out, "Panic recovered in goroutine")
	assert.Contains(t, out, "goroutine=\"test worker\"")
	assert.Contains(t, out, "panic=boom")
	// スタックトレースに panic した関数が含まれる
	assert.Contains(t, out, "TestRecoverGoroutine_RecoversPanicAndLogs")
}

func TestRecoverGoroutine_NoPanicLogsNothing(t *testing.T) {
	var buf bytes.Buffer
	logger := setupTestLogger(&buf)

	done := make(chan struct{})
	go func() {
		defer close(done)
		defer RecoverGoroutine(logger, "test worker")
	}()
	<-done

	assert.Empty(t, buf.String())
}

func TestRecoverGoroutine_WaitGroupDoesNotHang(t *testing.T) {
	var buf bytes.Buffer
	logger := setupTestLogger(&buf)

	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			defer RecoverGoroutine(logger, "test worker")
			if i == 1 {
				panic("boom")
			}
		}(i)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("wg.Wait hung after panic")
	}
}

func TestRecoverGoroutine_CallsOnPanic(t *testing.T) {
	var buf bytes.Buffer
	logger := setupTestLogger(&buf)

	got := make(chan any, 1)
	done := make(chan struct{})
	go func() {
		defer close(done)
		defer RecoverGoroutine(logger, "test worker", func(r any) { got <- r })
		panic("boom")
	}()
	<-done

	assert.Equal(t, "boom", <-got)
}

func TestRecoverGoroutine_NilLogger(t *testing.T) {
	done := make(chan struct{})
	go func() {
		defer close(done)
		defer RecoverGoroutine(nil, "test worker")
		panic("boom")
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("goroutine did not finish")
	}
}
