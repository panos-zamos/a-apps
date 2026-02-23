package templates

import (
	_ "embed"
	"net/http"
)

// customCSS is the shared stylesheet used by all apps.
//
//go:embed custom.css
var customCSS []byte

// CustomCSSHandler serves the embedded custom.css at runtime.
func CustomCSSHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(customCSS)
	}
}
