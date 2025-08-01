package ev

import (
	"cmp"
	"fmt"
	"os"
	"sync"
)

var (
	defaultValues map[string]any
	mu            sync.RWMutex
)

// SetDefault sets default value for the given var.
//
// This function is safe to call from multiple goroutines.
func SetDefault[T Constraint](v Var[T], value T) {
	mu.Lock()
	defer mu.Unlock()

	defaultValues[string(v)] = value
}

func getDefault[T Constraint](v Var[T]) T {
	mu.RLock()
	defer mu.RUnlock()

	value, ok := defaultValues[string(v)]
	if ok {
		return value.(T)
	}

	var zero T

	return zero
}

// Constraint defines all supported types for [Var].
//
// It does not permit ~T (tilde) types since other types may be parsed
// differently, which may result unexpected behavior.
type Constraint interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | uintptr | float32 | float64
}

// Var is an environment variable.
//
// T defines which value type this is expected by this variable.
type Var[T Constraint] string

// Or returns the first of its arguments for which [Var.TryGet] returns ok == true.
// If no argument is suitable, it returns the zero value.
func Or[T Constraint](vars ...Var[T]) T {
	for _, v := range vars {
		value, ok := v.TryGet()
		if ok {
			return value
		}
	}

	var zero T

	return zero
}

func (v Var[T]) get() string { return os.Getenv(string(v)) }

// GetOr returns the first of its arguments that is not equal to the zero value.
// If no argument is non-zero, it returns the zero value.
func (v Var[T]) GetOr(or ...T) T {
	value, ok := v.TryGet()
	if !ok {
		return cmp.Or(or...)
	}

	return value
}

// Get returns the value of this variable.
// It value is missing or not scannable by [fmt.Sscanf] it returns default value.
//
// See also [SetDefault] to set a default value.
func (v Var[T]) Get() T {
	value, ok := v.TryGet()
	if !ok {
		return getDefault(v)
	}

	return value
}

// TryGet returns a value of this variable and boolean
// stating whether value was present and [fmt.Sscanf] successfully scanned it.
func (v Var[T]) TryGet() (T, bool) {
	raw := v.get()
	if raw == "" {
		var zero T

		return zero, false
	}

	var out T

	_, err := fmt.Sscanf(raw, "%v", &out)
	if err != nil {
		var zero T

		return zero, false
	}

	return out, true
}
