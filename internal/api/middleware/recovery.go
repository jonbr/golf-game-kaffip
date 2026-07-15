package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

func RecoveryMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {

					// pull request-scoped logger if available
					reqLogger, ok := r.Context().Value("logger").(*slog.Logger)
					if !ok {
						reqLogger = logger
					}

					reqLogger.Error("panic recovered",
						"panic", rec,
						"stack", string(debug.Stack()),
					)

					http.Error(w, "internal server error", http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
