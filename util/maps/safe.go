package maps

import "sync"

// Safe is a generic, thread-safe map of keys to values.
type Safe[K comparable, V any] struct {
	m  map[K]V
	mu sync.RWMutex
}

// NewSafe returns a new safe map. It accepts an optional map to use as the
// initial underlying map or, if it is nil, a new one is created with the
// default capacity.
//
// NOTE:
// While providing an initial map may be useful to pre-allocate memory or to
// set the initial values, it is not safe to use it concurrently and should
// not be accessed or modified outside the safe map methods.
func NewSafe[K comparable, V any](m map[K]V) *Safe[K, V] {
	if m == nil {
		m = make(map[K]V)
	}
	return &Safe[K, V]{m: m}
}

// Len returns the number of key-value pairs in the map.
func (sm *Safe[K, V]) Len() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.m)
}

// Keys returns an unsorted slice of all keys in the map. The keys are shallow
// copies of the ones in the map.
func (sm *Safe[K, V]) Keys() []K {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return Keys(sm.m)
}

// Values returns a slice of all values in the map. The values are shallow
// copies of the ones in the map.
func (sm *Safe[K, V]) Values() []V {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return Values(sm.m)
}

// Get returns the value associated with the key and a boolean indicating
// whether the key was found.
func (sm *Safe[K, V]) Get(key K) (value V, ok bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	value, ok = sm.m[key]
	return
}

// Put adds the key-value pair to the set. If the key is already in the map,
// its value is updated.
func (sm *Safe[K, V]) Put(key K, value V) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.m[key] = value
}

// PutX adds the key-value pair only if the key is already in the map,
// returning a boolean indicating whether the key was found and the value
// updated.
//
// Example:
//
//	var set *Set[string, int]
//	set.PutX("a", 1) // false
//	set.Put("a", 1)
//	set.PutX("a", 2) // true
func (sm *Safe[K, V]) PutX(key K, value V) (ok bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	_, ok = sm.m[key]
	if ok {
		sm.m[key] = value
	}
	return
}

// PutNX adds the key-value pair only if the key is not already in the map,
// returning a boolean indicating whether the key was not found and the value
// added.
//
// Example:
//
//	var set *Set[string, int]
//	set.PutNX("a", 1) // true
//	set.PutNX("a", 2) // false
func (sm *Safe[K, V]) PutNX(key K, value V) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	_, ok := sm.m[key]
	if !ok {
		sm.m[key] = value
	}
	return !ok
}

// Pop returns the value associated with the key and a boolean indicating
// whether the key was found. If the key was found, it is removed from the map.
func (sm *Safe[K, V]) Pop(key K) (value V, ok bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	value, ok = sm.m[key]
	delete(sm.m, key)
	return
}

// Swap replaces the value associated with the key and returns the old value
// and a boolean indicating whether the key was found. If the key is not in the
// map, the value is still added and the result is the zero value of its type.
//
// Example:
//
//	var set *Set[string, int]
//	set.Put("a", 1)
//	set.Swap("a", 2) // 1, true
//	set.Swap("b", 3) // 0, false
func (sm *Safe[K, V]) Swap(key K, value V) (old V, ok bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	old, ok = sm.m[key]
	sm.m[key] = value
	return
}

// SwapX replaces the value associated with the key only if it is already in
// the map, returning a boolean indicating whether the key was found and the old
// value. If the key is not in the map, the value is not updated and the result
// is the zero value of its type.
//
// Example:
//
//	var set *Set[string, int]
//	set.Put("a", 1)
//	set.SwapX("a", 2) // 1, true
//	set.SwapX("b", 3) // 0, false
func (sm *Safe[K, V]) SwapX(key K, value V) (old V, ok bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	old, ok = sm.m[key]
	if ok {
		sm.m[key] = value
	}
	return
}

// Exist returns a boolean indicating whether the key is in the map.
//
// Example:
//
//	var set *Set[string, int]
//	set.Put("a", 1)
//	set.Exist("a") // true
//	set.Exist("b") // false
func (sm *Safe[K, V]) Exist(key K) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	_, ok := sm.m[key]
	return ok
}

// Delete removes the key-value pair from the set. If the key is not in the map,
// Delete is a no-op.
func (sm *Safe[K, V]) Delete(key K) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.m, key)
}

// DeleteX removes an existing the key-value pair from the map. It returns a
// boolean indicating whether the key was found and the value removed.
//
// Example:
//
//	var set *Set[string, int]
//	set.Put("a", 1)
//	set.DeleteX("a") // true
//	set.DeleteX("a") // false
func (sm *Safe[K, V]) DeleteX(key K) (ok bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	_, ok = sm.m[key]
	if ok {
		delete(sm.m, key)
	}
	return ok
}

// Clear removes all key-value pairs from the set.
func (sm *Safe[K, V]) Clear() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.m = make(map[K]V)
}

// ForEach calls the given function for each key-value pair in the map. The
// function receives the key and value as arguments. It returns the safe map
// to allow chaining.
//
// Example:
//
//	var set *Set[string, int]
//	set.Put("a", 1)
//	set.Put("b", 2)
//	set.ForEach(func(k string, v int) {
//		fmt.Println(k, v)
//		// "a 1"
//		// "b 2"
//	})
func (sm *Safe[K, V]) ForEach(fn func(K, V)) *Safe[K, V] {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	for k, v := range sm.m {
		fn(k, v)
	}
	return sm
}
