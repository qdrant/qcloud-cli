package framework

import "testing"

func TestJWTLikeRedaction(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "jwt token",
			input: "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.abc123_signature-value",
			want:  "[REDACTED]",
		},
		{
			name:  "jwt embedded in text",
			input: "Your key: eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.abc123_signature-value here",
			want:  "Your key: [REDACTED] here",
		},
		{
			name:  "no jwt",
			input: "just a normal string",
			want:  "just a normal string",
		},
		{
			name:  "dotted but too short segments",
			input: "a.b.c",
			want:  "a.b.c",
		},
		{
			name:  "version-like string not matched",
			input: "v1.2.3",
			want:  "v1.2.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := jwtLikeRe.ReplaceAllString(tt.input, "[REDACTED]")
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
