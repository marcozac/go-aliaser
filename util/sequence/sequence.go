package sequence

import (
	"slices"
)

// Sequence represents a type that can be expressed as a sequence of elements
// of type [T]. Under the hood, it uses a length function to get the
// number of elements and a getter to retrieve each one.
//
// Concurrent use of Sequence is safe as long as the provided length and getter
// functions are concurrent-safe, along with any operations performed on the
// sequence, such as those in the [Sequence.ForEach] method
type Sequence[T any] struct {
	len func() int
	at  func(int) T
}

// New returns a new [Sequence] with the given length and getter functions.
func New[T any](len func() int, at func(int) T) *Sequence[T] {
	return &Sequence[T]{len, at}
}

// Len is a wrapper for the length function.
func (seq *Sequence[T]) Len() int {
	return seq.len()
}

// At is a wrapper for the getter function.
func (seq *Sequence[T]) At(i int) T {
	return seq.at(i)
}

// Slice returns a new slice with the same length and the elements of the
// sequence.
func (seq *Sequence[T]) Slice() []T {
	slice := make([]T, seq.len())
	seq.ForEachIndex(func(e T, i int) {
		slice[i] = e
	})
	return slice
}

// SliceFunc returns a new slice with the results of calling the given function
// for each element in the sequence. It always returns a slice with the same
// length as the original sequence. If you want to exclude some elements, for
// example, zero values or in case of an error, use the [Sequence.SliceFuncFilter]
// method.
func (seq *Sequence[T]) SliceFunc(fn func(T) T) []T {
	slice := make([]T, seq.len())
	seq.ForEachIndex(func(e T, i int) {
		slice[i] = fn(e)
	})
	return slice
}

// SliceFuncIndex returns a new slice with the results of calling the given
// function for each element in the sequence, as [Sequence.SliceFunc], but also
// provides the index of the element, avoiding the need to use a closure to
// capture the index.
func (seq *Sequence[T]) SliceFuncIndex(fn func(T, int) T) []T {
	slice := make([]T, seq.len())
	seq.ForEachIndex(func(e T, i int) {
		slice[i] = fn(e, i)
	})
	return slice
}

// SliceFuncFilter returns a new slice with the results of calling the given
// function for each element in the sequence, whose result is appended to the
// slice only if the second value is true. The slice is clipped to remove
// unused capacity.
func (seq *Sequence[T]) SliceFuncFilter(fn func(T) (T, bool)) []T {
	slice := make([]T, 0, seq.len())
	seq.ForEach(func(e T) {
		if v, ok := fn(e); ok {
			slice = append(slice, v)
		}
	})
	return slices.Clip(slice)
}

// ForEach calls the given function for each element in the sequence.
func (seq *Sequence[T]) ForEach(fn func(T)) {
	for i := 0; i < seq.len(); i++ {
		fn(seq.At(i))
	}
}

// ForEachIndex calls the given function for each element in the sequence as
// [Sequence.ForEach], but also provides the index of the element. It avoids
// the need to use a closure to capture the index.
func (seq *Sequence[T]) ForEachIndex(fn func(T, int)) {
	for i := 0; i < seq.len(); i++ {
		fn(seq.At(i), i)
	}
}

// Sequenceable is a generic interface for types that can be expressed as a
// sequence of elements. It is used to provide a common interface for types
// that are not slices or arrays, but can be used as a sequence.
type Sequenceable[T any] interface {
	Len() int
	At(int) T
}

// FromSequenceable returns a new [Sequence] from the given value that
// implements the [Sequenceable] interface.
//
// Example:
//
//	import (
//		"go/types"
//
//		"github.com/marcozac/go-aliaser/util/sequence"
//	)
//
//	var tuple *types.Tuple
//	seq := sequence.FromSequenceable(tuple)
func FromSequenceable[T any](s Sequenceable[T]) *Sequence[T] {
	return New(s.Len, s.At)
}
