package xsync

import (
	"sync"
)

type SlicePool[T any] interface {
	Get() []T
	Put([]T)
}

type BatchSliceConfig struct {
	MaxBufferCap int // max capacity after batch operations before shrinking
}

// BatchSlice is a thread-safe, high-performance slice buffer with batching capabilities
type BatchSlice[T any] struct {
	mu     sync.Mutex
	slice  []T
	pool   SlicePool[T]
	config *BatchSliceConfig
}

func NewBatchSlice[T any](startCap int, pool SlicePool[T], config *BatchSliceConfig) *BatchSlice[T] {
	return &BatchSlice[T]{
		slice:  make([]T, 0, startCap),
		pool:   pool,
		config: config,
	}
}

func (s *BatchSlice[T]) Append(rows ...T) {
	s.mu.Lock()
	s.slice = append(s.slice, rows...)
	s.mu.Unlock()
}

func (s *BatchSlice[T]) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.slice)
}

func (s *BatchSlice[T]) Clear() {
	s.mu.Lock()
	s.slice = nil
	s.mu.Unlock()
}

// PeekBatch returns up to max oldest rows without removing them
func (s *BatchSlice[T]) PeekBatch(max int) []T {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.slice) == 0 {
		return nil
	}

	n := min(len(s.slice), max)
	return s.slice[:n]
}

// PeekBatchCopy returns up to max oldest rows as a copy
func (s *BatchSlice[T]) PeekBatchCopy(max int) []T {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.slice) == 0 {
		return nil
	}

	n := min(len(s.slice), max)
	return append([]T(nil), s.slice[:n]...)
}

// TakeBatch removes up to max oldest rows and returns them
func (s *BatchSlice[T]) TakeBatch(max int) []T {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.slice) == 0 {
		return nil
	}

	n := min(len(s.slice), max)

	batch := s.slice[:n]
	s.slice = s.shrinkIfNeeded(s.slice[n:])
	return batch
}

// TakeBatchCopy removes up to max oldest rows and returns a copy
func (s *BatchSlice[T]) TakeBatchCopy(max int) []T {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.slice) == 0 {
		return nil
	}

	n := min(len(s.slice), max)

	batch := append([]T(nil), s.slice[:n]...)
	s.slice = s.shrinkIfNeeded(s.slice[n:])
	return batch
}

// TakeBatchAndRelease removes up to max rows and releases the slice immediately
func (s *BatchSlice[T]) TakeBatchAndRelease(max int) []T {
	batch := s.TakeBatch(max)
	if batch != nil {
		s.Release(batch)
	}
	return batch
}

// TakeBatchWithRelease removes batch, executes fn, and releases the slice to the pool
func (s *BatchSlice[T]) TakeBatchWithRelease(max int, fn func(batch []T)) {
	batch := s.TakeBatch(max)
	if batch == nil {
		return
	}
	fn(batch)
	s.Release(batch)
}

// CommitN removes the first n rows
func (s *BatchSlice[T]) Commit(n int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if n >= len(s.slice) {
		s.slice = nil
		return
	}
	s.slice = s.shrinkIfNeeded(s.slice[n:])
}

// CommitNAndRelease removes the first n rows and releases them to the pool
func (s *BatchSlice[T]) CommitAndRelease(n int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if n >= len(s.slice) {
		if s.pool != nil {
			s.pool.Put(s.slice)
		}
		s.slice = nil
		return
	}

	batch := s.slice[:n]
	s.slice = s.shrinkIfNeeded(s.slice[n:])
	s.Release(batch)
}

// Release puts a slice back into the pool if pool is defined
func (s *BatchSlice[T]) Release(rows []T) {
	if s.pool != nil && rows != nil {
		s.pool.Put(rows)
	}
}

// Front returns the oldest row without removing it
func (s *BatchSlice[T]) Front() (T, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.slice) == 0 {
		var zero T
		return zero, false
	}
	return s.slice[0], true
}

// Back returns the newest row without removing it
func (s *BatchSlice[T]) Back() (T, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.slice) == 0 {
		var zero T
		return zero, false
	}
	return s.slice[len(s.slice)-1], true
}

// shrinkIfNeeded checks MaxBufferCap and optionally reallocates the slice
func (s *BatchSlice[T]) shrinkIfNeeded(remaining []T) []T {
	if s.config != nil && s.config.MaxBufferCap > 0 && cap(remaining) > s.config.MaxBufferCap {
		if s.pool != nil {
			newSlice := s.pool.Get()[:0]
			newSlice = append(newSlice, remaining...)
			return newSlice
		}
		return append([]T(nil), remaining...)
	}
	return remaining
}
