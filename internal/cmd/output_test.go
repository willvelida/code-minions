package cmd

import "testing"

func TestNonNil(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  int // expected length (0 for empty, >0 for non-nil passthrough)
	}{
		{name: "nil returns empty slice", input: nil, want: 0},
		{name: "empty returns empty", input: []string{}, want: 0},
		{name: "populated returns same", input: []string{"a", "b"}, want: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := nonNil(tt.input)
			if got == nil {
				t.Fatal("nonNil should never return nil")
			}
			if len(got) != tt.want {
				t.Errorf("len = %d, want %d", len(got), tt.want)
			}
		})
	}
}

func TestSortedKeys(t *testing.T) {
	tests := []struct {
		name string
		m    map[string]int
		want []string
	}{
		{name: "nil map", m: nil, want: []string{}},
		{name: "empty map", m: map[string]int{}, want: []string{}},
		{name: "single key", m: map[string]int{"a": 1}, want: []string{"a"}},
		{name: "sorted output", m: map[string]int{"c": 3, "a": 1, "b": 2}, want: []string{"a", "b", "c"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sortedKeys(tt.m)
			if len(got) != len(tt.want) {
				t.Fatalf("len = %d, want %d", len(got), len(tt.want))
			}
			for i, k := range got {
				if k != tt.want[i] {
					t.Errorf("keys[%d] = %q, want %q", i, k, tt.want[i])
				}
			}
		})
	}
}
