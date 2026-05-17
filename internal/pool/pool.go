// internal/pool/pool.go
package pool

import (
	"sync"
)

type Resetter interface {
	Reset()
}

type Pool[T any] struct {
	pool sync.Pool

	newFunc func() T
}

func New[T Resetter](newFunc func() T) *Pool[T] {
	p := &Pool[T]{
		newFunc: newFunc,
	}

	p.pool.New = func() any {
		return p.newFunc()
	}

	return p
}

func (p *Pool[T]) Get() T {
	if obj := p.pool.Get(); obj != nil {
		return obj.(T)
	}

	return p.newFunc()
}

func (p *Pool[T]) Put(obj T) {
	if resetter, ok := any(obj).(Resetter); ok {
		resetter.Reset()
	}

	p.pool.Put(obj)
}
