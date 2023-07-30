package structures

import (
	"math"
	"math/rand"
	"sync"
)

type Pool[T any] interface {
	Put(...T)
	Take(int) []T
	Length() int
	Shuffle()
}

type pool[T any] struct {
	values []T
	mutex  *sync.Mutex
}

func (p *pool[T]) Put(items ...T) {
	p.mutex.Lock()
	p.values = append(p.values, items...)
	p.mutex.Unlock()
}

func (p *pool[T]) Take(count int) []T {
	p.mutex.Lock()

	length := len(p.values)
	endIndex := int(math.Min(float64(count), float64(length)))
	currentBatch := p.values[:endIndex]
	p.values = p.values[endIndex:]

	p.mutex.Unlock()

	return currentBatch
}

func (p *pool[T]) Shuffle() {
	p.mutex.Lock()
	for i := range p.values {
		j := rand.Intn(i + 1)
		p.values[i], p.values[j] = p.values[j], p.values[i]
	}
	p.mutex.Unlock()
}

func (p pool[T]) Length() int {
	return len(p.values)
}

func CreatePool[T any](items []T) Pool[T] {
	return &pool[T]{
		values: items,
		mutex:  &sync.Mutex{},
	}
}
