package reference

import "testing"

func TestNumber_Next(t *testing.T) {
	tests := []struct {
		prev Number
		next Number
	}{
		{"123443", "123453"},
		{"13504674", "13504687"},
	}
	for _, tt := range tests {
		t.Run(string(tt.prev), func(t *testing.T) {
			if got := tt.prev.Next(); got != tt.next {
				t.Errorf("Next(%v) = %v, want %v", tt.prev, got, tt.next)
			}
		})
	}
}
