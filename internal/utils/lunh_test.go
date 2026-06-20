package utils

import "testing"

func TestIsValidLuhn(t *testing.T) {
	tests := []struct {
		name   string
		number string
		want   bool
	}{
		{"valid", "4532015112830366", true},
		{"invalid", "4532015112830367", false},
		{"empty", "", false},
		{"non-digit", "123a", false},
		{"short", "12", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidLuhn(tt.number); got != tt.want {
				t.Errorf("IsValidLuhn(%s) = %v, want %v", tt.number, got, tt.want)
			}
		})
	}
}
