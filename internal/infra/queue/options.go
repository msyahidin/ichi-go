package queue

import "time"

// DispatchOptions holds resolved values after applying all DispatchOption funcs.
type DispatchOptions struct {
	Queue       string
	Delay       time.Duration
	MaxAttempts int
	Priority    int
	UniqueKey   string
}

// DispatchOption mutates DispatchOptions.
type DispatchOption func(*DispatchOptions)

// ApplyOptions builds a DispatchOptions with defaults, then applies each opt.
func ApplyOptions(opts ...DispatchOption) *DispatchOptions {
	o := &DispatchOptions{
		Queue:       "default",
		MaxAttempts: 3,
		Priority:    1,
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

func OnQueue(name string) DispatchOption {
	return func(o *DispatchOptions) { o.Queue = name }
}

func Delay(d time.Duration) DispatchOption {
	return func(o *DispatchOptions) { o.Delay = d }
}

func MaxAttempts(n int) DispatchOption {
	return func(o *DispatchOptions) { o.MaxAttempts = n }
}

func Priority(p int) DispatchOption {
	return func(o *DispatchOptions) { o.Priority = p }
}

func UniqueKey(k string) DispatchOption {
	return func(o *DispatchOptions) { o.UniqueKey = k }
}
