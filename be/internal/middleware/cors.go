package middleware

import (
	"net/http"

	"github.com/go-chi/cors"
)

func CORS(allowedOrigin string) func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins: []string{allowedOrigin},
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{
			"Content-Type",
			"Authorization",
			"Cookie", // add this
			"X-CSRF-Token",
		},
		ExposedHeaders:   []string{"Set-Cookie"}, // add this
		AllowCredentials: true,
		MaxAge:           300,
	})
}
