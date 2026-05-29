package logger

import (
	"regexp"
	"strings"
)

var cardPattern = regexp.MustCompile(`\b(?:\d[ -]*?){13,19}\b`)

type SensitiveDataSanitizer struct{}

func NewSanitizer() *SensitiveDataSanitizer {
	return &SensitiveDataSanitizer{}
}

func (s *SensitiveDataSanitizer) SanitizeMap(data map[string]any) map[string]any {
	if data == nil {
		return nil
	}
	out := make(map[string]any, len(data))
	for k, v := range data {
		out[k] = s.sanitizeValue(k, v)
	}
	return out
}

func (s *SensitiveDataSanitizer) sanitizeValue(key string, value any) any {
	if isSensitiveKey(key) {
		return "***MASKED***"
	}
	switch v := value.(type) {
	case map[string]any:
		return s.SanitizeMap(v)
	case []any:
		arr := make([]any, len(v))
		for i, item := range v {
			arr[i] = s.sanitizeValue(key, item)
		}
		return arr
	case string:
		if cardPattern.MatchString(v) {
			return "***MASKED***"
		}
		return v
	default:
		return value
	}
}

func isSensitiveKey(key string) bool {
	k := strings.ToLower(key)
	sensitive := []string{
		"password", "passwd", "token", "authorization", "auth",
		"secret", "api_key", "apikey", "cvv", "card", "credit_card",
		"credit card", "pan",
	}
	for _, s := range sensitive {
		if strings.Contains(k, s) {
			return true
		}
	}
	return false
}
