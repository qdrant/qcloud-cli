package testutil

import (
	"context"
	"sync"
)

// MethodSpy records calls to a single RPC method and optionally dispatches
// to per-call or fallback handlers. Use the *Calls fields on fake services
// instead of managing call counters and captured requests manually.
//
// Priority when handling a call: *Func field > OnCall handler > Always handler > Unimplemented default.
type MethodSpy[Req, Resp any] struct {
	mu       sync.Mutex
	requests []Req
	handlers []func(context.Context, Req) (Resp, error)
	fallback func(context.Context, Req) (Resp, error)
}

// OnCall registers a handler for the i-th call (0-indexed). Chainable.
func (m *MethodSpy[Req, Resp]) OnCall(i int, fn func(context.Context, Req) (Resp, error)) *MethodSpy[Req, Resp] {
	m.mu.Lock()
	defer m.mu.Unlock()
	for len(m.handlers) <= i {
		m.handlers = append(m.handlers, nil)
	}
	m.handlers[i] = fn
	return m
}

// Returns sets a fixed response returned for every call (shorthand for Always with a constant return).
func (m *MethodSpy[Req, Resp]) Returns(resp Resp, err error) *MethodSpy[Req, Resp] {
	return m.Always(func(_ context.Context, _ Req) (Resp, error) {
		return resp, err
	})
}

// Always sets a fallback handler invoked for every call that has no matching OnCall entry.
// If neither OnCall nor Always is set the method falls back to the gRPC Unimplemented default.
func (m *MethodSpy[Req, Resp]) Always(fn func(context.Context, Req) (Resp, error)) *MethodSpy[Req, Resp] {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.fallback = fn
	return m
}

// Count returns how many times the method has been called.
func (m *MethodSpy[Req, Resp]) Count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.requests)
}

// All returns a copy of all captured requests in call order.
func (m *MethodSpy[Req, Resp]) All() []Req {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Req, len(m.requests))
	copy(out, m.requests)
	return out
}

// Last returns the most recent request and true, or the zero value and false if never called.
func (m *MethodSpy[Req, Resp]) Last() (Req, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.requests) == 0 {
		var zero Req
		return zero, false
	}
	return m.requests[len(m.requests)-1], true
}

// Reset clears captured requests and all registered handlers.
func (m *MethodSpy[Req, Resp]) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requests = nil
	m.handlers = nil
	m.fallback = nil
}

// record appends the request to the captured list. Called before dispatch.
func (m *MethodSpy[Req, Resp]) record(req Req) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requests = append(m.requests, req)
}

// dispatch selects the handler for the current call index (based on len(requests)
// after recording), falls back to Always, then falls back to unimpl.
func (m *MethodSpy[Req, Resp]) dispatch(ctx context.Context, req Req, unimpl func(context.Context, Req) (Resp, error)) (Resp, error) {
	m.mu.Lock()
	idx := len(m.requests) - 1
	var handler func(context.Context, Req) (Resp, error)
	if idx >= 0 && idx < len(m.handlers) {
		handler = m.handlers[idx]
	}
	if handler == nil {
		handler = m.fallback
	}
	m.mu.Unlock()

	if handler != nil {
		return handler(ctx, req)
	}
	return unimpl(ctx, req)
}
