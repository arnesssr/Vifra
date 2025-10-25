package security

import (
	"net/http"
)

// SecurityHeadersMiddleware adds security headers to HTTP responses
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add security headers
		headers := w.Header()
		
		// Prevent MIME type sniffing
		headers.Set("X-Content-Type-Options", "nosniff")
		
		// Prevent clickjacking
		headers.Set("X-Frame-Options", "DENY")
		
		// XSS protection
		headers.Set("X-XSS-Protection", "1; mode=block")
		
		// Strict Transport Security (only if HTTPS is used)
		// headers.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		
		// Content Security Policy
		headers.Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:;")
		
		// Referrer Policy
		headers.Set("Referrer-Policy", "no-referrer")
		
		next.ServeHTTP(w, r)
	})
}

// CORSMiddleware adds CORS headers to HTTP responses
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add CORS headers
		headers := w.Header()
		
		// Allow requests from specific origins (configure as needed)
		headers.Set("Access-Control-Allow-Origin", "*") // In production, specify exact origins
		
		// Allow specific methods
		headers.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		
		// Allow specific headers
		headers.Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		
		// Allow credentials
		headers.Set("Access-Control-Allow-Credentials", "true")
		
		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}