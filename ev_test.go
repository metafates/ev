package ev

import (
	"testing"
)

func resetDefaults() {
	mu.Lock()
	defaultValues = make(map[string]any)
	mu.Unlock()
}

func TestVar_Get(t *testing.T) {
	resetDefaults()

	cases := map[string]Runner{
		"int/valid":                GetCase[int]{Value: "42", Want: 42},
		"int/with_default":         GetCase[int]{SetDefault: true, Default: 100, Want: 100},
		"int/without_default":      GetCase[int]{Want: 0},
		"int/invalid_no_default":   GetCase[int]{Value: "abc", Want: 0},
		"int/invalid_with_default": GetCase[int]{Value: "abc", SetDefault: true, Default: 100, Want: 100},
		"int/negative":             GetCase[int]{Value: "-42", Want: -42},
		"int/default_override":     GetCase[int]{Value: "42", SetDefault: true, Default: 100, Want: 42},

		"uint/valid":           GetCase[uint]{Value: "42", Want: 42},
		"uint/with_default":    GetCase[uint]{SetDefault: true, Default: 100, Want: 100},
		"uint/without_default": GetCase[uint]{Want: 0},
		"uint/negative":        GetCase[uint]{Value: "-42", Want: 0},

		"float32/valid":    GetCase[float32]{Value: "3.14", Want: 3.14},
		"float64/valid":    GetCase[float64]{Value: "6.283185", Want: 6.283185},
		"float64/negative": GetCase[float64]{Value: "-2.718", Want: -2.718},
	}

	for name, tc := range cases {
		t.Run(name, tc.Run)
	}
}

func TestVar_TryGet(t *testing.T) {
	cases := map[string]Runner{
		"valid":       TryGetCase[int]{Value: "42", Want: 42, Ok: true},
		"empty":       TryGetCase[int]{Value: "", Want: 0, Ok: false},
		"invalid":     TryGetCase[int]{Value: "abc", Want: 0, Ok: false},
		"uint_neg":    TryGetCase[uint]{Value: "-1", Want: 0, Ok: false},
		"float_valid": TryGetCase[float32]{Value: "1.618", Want: 1.618, Ok: true},
	}

	for name, tc := range cases {
		t.Run(name, tc.Run)
	}
}

func TestVar_GetOr(t *testing.T) {
	cases := map[string]Runner{
		"valid":          GetOrCase[int]{Value: "42", Or: []int{10, 20}, Want: 42},
		"empty":          GetOrCase[int]{Value: "", Or: []int{10, 20}, Want: 10},
		"invalid":        GetOrCase[int]{Value: "abc", Or: []int{10, 20}, Want: 10},
		"no_fallback":    GetOrCase[int]{Value: "abc", Want: 0},
		"multi_fallback": GetOrCase[int]{Value: "", Or: []int{1, 2, 3}, Want: 1},
		"float":          GetOrCase[float64]{Value: "3.14", Or: []float64{1.1, 2.2}, Want: 3.14},
	}

	for name, tc := range cases {
		t.Run(name, tc.Run)
	}
}

func TestVar_Or(t *testing.T) {
	cases := map[string]Runner{
		"first_set":      OrCase[int]{EnvVars: []string{"A", "B", "C"}, Values: []string{"10", "20", "30"}, Want: 10},
		"second_set":     OrCase[int]{EnvVars: []string{"D", "E", "F"}, Values: []string{"", "20", "30"}, Want: 20},
		"last_set":       OrCase[int]{EnvVars: []string{"G", "H", "I"}, Values: []string{"", "", "30"}, Want: 30},
		"none_set":       OrCase[int]{EnvVars: []string{"J", "K", "L"}, Values: []string{"", "", ""}, Want: 0},
		"first_invalid":  OrCase[int]{EnvVars: []string{"M", "N", "O"}, Values: []string{"abc", "20", "30"}, Want: 20},
		"middle_invalid": OrCase[int]{EnvVars: []string{"P", "Q", "R"}, Values: []string{"", "invalid", "30"}, Want: 30},
		"all_invalid":    OrCase[int]{EnvVars: []string{"S", "T", "U"}, Values: []string{"x", "y", "z"}, Want: 0},
		"float":          OrCase[float64]{EnvVars: []string{"V", "W", "X"}, Values: []string{"", "3.14", ""}, Want: 3.14},
		"mixed_types":    OrCase[uint]{EnvVars: []string{"Y", "Z"}, Values: []string{"", "255"}, Want: 255},
	}

	for name, tc := range cases {
		t.Run(name, tc.Run)
	}
}

// Test case implementations remain the same as before
type GetCase[T Constraint] struct {
	Value      string
	Want       T
	Default    T
	SetDefault bool
}

func (c GetCase[T]) Run(t *testing.T) {
	t.Cleanup(resetDefaults)

	const v Var[T] = "GET_CASE_TEST"
	t.Setenv(string(v), c.Value)

	if c.SetDefault {
		SetDefault(v, c.Default)
	}

	got := v.Get()
	if got != c.Want {
		t.Errorf("Var[%T].Get() = %v, want %v", got, got, c.Want)
	}
}

type TryGetCase[T Constraint] struct {
	Value string
	Want  T
	Ok    bool
}

func (c TryGetCase[T]) Run(t *testing.T) {
	const v Var[T] = "TRY_GET_CASE"
	t.Setenv(string(v), c.Value)

	got, ok := v.TryGet()
	if got != c.Want || ok != c.Ok {
		t.Errorf("Var[%T].TryGet() = (%v, %t), want (%v, %t)", got, got, ok, c.Want, c.Ok)
	}
}

type GetOrCase[T Constraint] struct {
	Value string
	Or    []T
	Want  T
}

func (c GetOrCase[T]) Run(t *testing.T) {
	const v Var[T] = "GET_OR_CASE"
	t.Setenv(string(v), c.Value)

	got := v.GetOr(c.Or...)
	if got != c.Want {
		t.Errorf("Var[%T].GetOr() = %v, want %v", got, got, c.Want)
	}
}

type OrCase[T Constraint] struct {
	EnvVars []string
	Values  []string
	Want    T
}

func (c OrCase[T]) Run(t *testing.T) {
	vars := make([]Var[T], len(c.EnvVars))
	for i, name := range c.EnvVars {
		vars[i] = Var[T](name)
		if i < len(c.Values) {
			t.Setenv(name, c.Values[i])
		}
	}

	got := Or(vars...)
	if got != c.Want {
		t.Errorf("Or(%v) = %v, want %v", c.EnvVars, got, c.Want)
	}
}

type Runner interface {
	Run(t *testing.T)
}
