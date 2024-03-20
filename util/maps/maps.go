package maps

// Keys creates and returns a new slice containing all the keys of the given
// map.
//
// NOTE:
// The keys are shallow copies of the keys in the map. If they contain any
// reference, the returned slice will hold references to the same objects as
// those in the map.
func Keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Values creates and returns a new slice containing all the values of the given
// map.
//
// NOTE:
// The values are shallow copies of the values in the map. If they contain any
// reference, the returned slice will hold references to the same objects as
// those in the map.
func Values[K comparable, V any](m map[K]V) []V {
	values := make([]V, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}
