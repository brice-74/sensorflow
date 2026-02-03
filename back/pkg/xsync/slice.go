package xsync

import (
	"sync"
)

type SlicePool[T any] interface {
	Get() []T
	Put([]T)
}

// BatchSlice is a thread-safe, memory-efficient slice buffer with built-in batching.
//
// Features:
// - Batch extraction: TakeBatch(max) returns up to max elements
// - Optional slice pooling for reduced allocations and GC pressure
// - Safe for high-throughput, concurrent use
type BatchSlice[T any] struct {
	mu    sync.Mutex
	slice []T
	pool  SlicePool[T] // optional: to reuse slices and reduce GC pressure
}

func NewBatchSlice[T any](startCap int, pool SlicePool[T]) *BatchSlice[T] {
	return &BatchSlice[T]{
		slice: make([]T, 0, startCap),
		pool:  pool,
	}
}

// Append rows to the active buffer
func (s *BatchSlice[T]) Append(rows []T) {
	s.mu.Lock()
	s.slice = append(s.slice, rows...)
	s.mu.Unlock()
}

// TakeBatch retrieves up to max elements for flushing.
// Returns a separate slice and clears the active buffer if necessary.
func (s *BatchSlice[T]) TakeBatch(max int) []T {
	s.mu.Lock()
	defer s.mu.Unlock()

	n := len(s.slice)
	if n == 0 {
		return nil
	}

	if n > max {
		batch := s.slice[:max]

		remaining := s.slice[max:]
		if s.pool != nil {
			newSlice := s.pool.Get()[:0]
			newSlice = append(newSlice, remaining...)
			s.slice = newSlice
		} else {
			s.slice = append([]T(nil), remaining...)
		}

		return batch
	}

	batch := s.slice
	if s.pool != nil {
		s.slice = s.pool.Get()[:0]
	} else {
		s.slice = nil
	}
	return batch
}

func (s *BatchSlice[T]) TakeBatchWithRelease(max int, fn func(batch []T)) {
	batch := s.TakeBatch(max)
	if batch == nil {
		return
	}
	fn(batch)
	s.Release(batch)
}

func (s *BatchSlice[T]) Release(rows []T) {
	if s.pool != nil {
		s.pool.Put(rows)
	}
}

// Len returns the number of rows in the active buffer
func (s *BatchSlice[T]) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.slice)
}
