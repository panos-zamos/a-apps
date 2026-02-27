package auth

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const UserContextKey contextKey = "username"

// Middleware creates an authentication middleware
func Middleware(jwtSecret, basePath string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check for JWT cookie
			cookie, err := r.Cookie("auth_token")
			if err != nil {
				// No cookie, redirect to login
				http.Redirect(w, r, prefixPath(basePath, "/login"), http.StatusSeeOther)
				return
			}

			// Validate token
			username, err := ValidateToken(cookie.Value, jwtSecret)
			if err != nil {
				// Invalid token, redirect to login
				http.Redirect(w, r, prefixPath(basePath, "/login"), http.StatusSeeOther)
				return
			}

			// Add username to context
			ctx := context.WithValue(r.Context(), UserContextKey, username)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func prefixPath(basePath, path string) string {
	basePath = strings.TrimSpace(basePath)
	if basePath == "" || basePath == "/" {
		return path
	}
	if !strings.HasPrefix(basePath, "/") {
		basePath = "/" + basePath
	}
	basePath = strings.TrimRight(basePath, "/")
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return basePath + path
}

// GetUsername extracts username from request context
func GetUsername(r *http.Request) (string, bool) {
	username, ok := r.Context().Value(UserContextKey).(string)
	return username, ok
}
