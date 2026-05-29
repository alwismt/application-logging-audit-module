package logger

import "testing"

func TestSensitiveDataSanitizer_SanitizeMap(t *testing.T) {
	s := NewSanitizer()
	tests := []struct {
		name string
		in   map[string]any
		key  string
	}{
		{"password", map[string]any{"password": "secret123"}, "password"},
		{"token", map[string]any{"access_token": "abc"}, "access_token"},
		{"authorization", map[string]any{"authorization": "Bearer x"}, "authorization"},
		{"secret", map[string]any{"api_secret": "key"}, "api_secret"},
		{"card", map[string]any{"credit_card": "4111111111111111"}, "credit_card"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := s.SanitizeMap(tt.in)
			if out[tt.key] != "***MASKED***" {
				t.Fatalf("expected masked value for %s, got %v", tt.key, out[tt.key])
			}
		})
	}
}

func TestSensitiveDataSanitizer_CardNumberInString(t *testing.T) {
	s := NewSanitizer()
	out := s.SanitizeMap(map[string]any{"note": "paid with 4111111111111111"})
	if out["note"] != "***MASKED***" {
		t.Fatalf("expected card number masked, got %v", out["note"])
	}
}
