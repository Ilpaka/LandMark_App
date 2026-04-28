package authhttp

import (
	"testing"
)

func TestValidPassword(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   string
		want bool
	}{
		{"short1", false},
		{"alllettersno", false},
		{"12345678901", false},
		{"GoodPassw0rd", true},
		{"ten_chars1", true},
	}
	for _, tc := range cases {
		if got := validPassword(tc.in); got != tc.want {
			t.Fatalf("validPassword(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}
