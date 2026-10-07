package output

import (
	"fmt"
	"io"
)

// KeySecret prints the one-time secret of a newly created key, if present.
func KeySecret(out io.Writer, secret string) {
	if secret == "" {
		return
	}

	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Save this key now — it will not be shown again:")
	fmt.Fprintf(out, "  %s\n", secret)
}

// CreatedAPIKey prints the creation message and the one-time secret of an API key.
func CreatedAPIKey(out io.Writer, key interface {
	GetId() string
	GetName() string
	GetKey() string
},
) {
	fmt.Fprintf(out, "API key %s (%s) created.\n", key.GetId(), key.GetName())
	KeySecret(out, key.GetKey())
}
