package codegen

import "testing"

func TestModelName(t *testing.T) {
	tests := []struct {
		table string
		want  string
	}{
		{"users", "Users"},
		{"gas", "Gas"},
		{"news", "News"},
		{"user_status", "UserStatus"},
	}
	for _, tt := range tests {
		if got := modelName(tt.table); got != tt.want {
			t.Errorf("modelName(%q) = %q, want %q", tt.table, got, tt.want)
		}
	}
}
