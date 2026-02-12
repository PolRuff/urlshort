package pool

import "sync"

// Resettable — ограничение для generic-параметра.
// Любой тип, который можно положить в Pool, обязан
// иметь метод Reset().
type Resettable interface {
	Reset()
}

// Pool — generic-контейнер для объектов одного типа.
type Pool[T Resettable] struct {
	pool sync.Pool
}

// New — конструктор Pool.
// newFn используется для создания нового объекта,
// если пул пуст.
func New[T Resettable](newFn func() T) *Pool[T] {
	return &Pool[T]{
		pool: sync.Pool{
			New: func() any {
				return newFn()
			},
		},
	}
}

// Get — получает объект из пула.
func (p *Pool[T]) Get() T {
	return p.pool.Get().(T)
}

// Put — возвращает объект в пул.
// Перед возвратом обязательно сбрасываем состояние.
func (p *Pool[T]) Put(obj T) {
	obj.Reset()
	p.pool.Put(obj)
}
