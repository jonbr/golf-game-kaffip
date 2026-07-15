package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

func RequestIDMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// 1. Generate request ID
			id := uuid.New().String()

			// 2. Create a request-scoped logger
			reqLogger := logger.With("request_id", id)

			// 3. Store both ID and logger in context
			ctx := context.WithValue(r.Context(), "request_id", id)
			ctx = context.WithValue(ctx, "logger", reqLogger)

			// 4. Add header for clients
			w.Header().Set("X-Request-ID", id)

			// 5. Continue request with enriched context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
