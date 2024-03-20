// Copyright (c) 2023 ISK SRL. All rights reserved.

package set

import "sync"

// Set is a thread-safe map of keys to values.
type Set[K comparable, V any] struct {
	mu sync.RWMutex
	sm map[K]V
}

// NewSet returns a new set of keys to values with the given type parameters.
func New[K comparable, V any]() *Set[K, V] {
	return &Set[K, V]{
		sm: make(map[K]V),
	}
}

// Get returns the value associated with the key and a boolean indicating
// whether the key was found.
func (s *Set[K, V]) Get(key K) (V, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.sm[key]
	return v, ok
}

// Pop returns the value associated with the key and a boolean indicating
// whether the key was found. If the key was found, it is removed from the set.
func (s *Set[K, V]) Pop(key K) (V, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.sm[key]
	if ok {
		delete(s.sm, key)
	}
	return v, ok
}

// Put adds the key-value pair to the set. If the key is already in the set,
// its value is updated.
func (s *Set[K, V]) Put(key K, value V) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sm[key] = value
}

// PutX adds the key-value pair only if the key is already in the set,
// returning a boolean indicating whether the key was found and the value
// updated.
func (s *Set[K, V]) PutX(key K, value V) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.sm[key]
	if ok {
		s.sm[key] = value
	}
	return ok
}

// PutNX adds the key-value pair only if the key is not already in the set,
// returning a boolean indicating whether the key was not found and the value
// added.
func (s *Set[K, V]) PutNX(key K, value V) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.sm[key]
	if !ok {
		s.sm[key] = value
	}
	return !ok
}

// Swap replaces the value associated with the key and returns the old value.
// If the key is not in the set, the value is still added and the result is the
// zero value of the value type.
func (s *Set[K, V]) Swap(key K, value V) V {
	s.mu.Lock()
	defer s.mu.Unlock()
	old := s.sm[key]
	s.sm[key] = value
	return old
}

// Delete removes the key-value pair from the set. It returns a boolean
// indicating whether the key was found and the value removed.
func (s *Set[K, V]) Delete(key K) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.sm[key]
	if ok {
		delete(s.sm, key)
	}
	return ok
}
