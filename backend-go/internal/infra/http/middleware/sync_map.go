// Package middleware реализует middleware
package middleware

import "sync"

type typedSyncMap[K comparable, V any] struct {
	m sync.Map
}

func (s *typedSyncMap[K, V]) Load(key K) (V, bool) {
	v, exists := s.m.Load(key)
	if !exists {
		var zero V
		return zero, false
	}
	val, ok := v.(V)
	if !ok {
		var zero V
		return zero, false
	}
	return val, ok
}

func (s *typedSyncMap[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	a, loaded := s.m.LoadOrStore(key, value)
	val, ok := a.(V)
	if !ok {
		var zero V
		return zero, false
	}
	return val, loaded
}
