package oauth

import (
	_ "embed"
	"html/template"
	"net/http"
)

//go:embed callback.html
var callbackPageSource string

// callbackPage is what the browser shows on the loopback redirect. It is self-contained: the logos
// are inline SVG and there is no font, script or other request, so it renders offline too.
var callbackPage = template.Must(template.New("callback").Parse(callbackPageSource))

// writeCallbackPage renders the success page, or the failure page with detail (HTML-escaped).
func writeCallbackPage(w http.ResponseWriter, status int, detail string) {
	h := w.Header()
	h.Set("Content-Type", "text/html; charset=utf-8")
	h.Set("Cache-Control", "no-store")
	h.Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; img-src data:; base-uri 'none'; form-action 'none'; frame-ancestors 'none'")
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("Referrer-Policy", "no-referrer")
	w.WriteHeader(status)

	_ = callbackPage.Execute(w, struct {
		OK     bool
		Detail string
	}{OK: status == http.StatusOK, Detail: detail})
}
