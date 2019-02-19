package calc

import "testing"

func TestMin(t *testing.T) {
	testCases := []struct {
		a        int64
		b        int64
		expected int64
	}{
		{a: 1, b: 4, expected: 1},
		{a: 3, b: 2, expected: 2},
	}
	for i, tc := range testCases {
		if s := Min(tc.a, tc.b); s != tc.expected {
			t.Errorf("[%d] Expecting %v, got %v", i+1, tc.expected, s)
		}
	}
}
