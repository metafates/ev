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

type Constraint interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr | ~float32 | ~float64
}

type Var[T Constraint] string

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

func (v Var[T]) GetOr(or ...T) T {
	value, ok := v.TryGet()
	if !ok {
		return cmp.Or(or...)
	}

	return value
}

func (v Var[T]) Get() T {
	value, ok := v.TryGet()
	if !ok {
		return getDefault(v)
	}

	return value
}

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
