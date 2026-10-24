package api

import (
	"context"
	"golf-game-kaffip/internal/logging"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func StartRequest(r *http.Request, action string) (context.Context, *slog.Logger) {
	ctx := r.Context()
	logger := logging.FromCtx(ctx)
	logger.Info(action)
	return ctx, logger
}

// TODO: Make function agnostic so we can parse other Id's as well.
func ParseID(w http.ResponseWriter, r *http.Request, logger *slog.Logger, entityName string) (string, bool) {
	id := chi.URLParam(r, "id")
	if id == "" {
		logger.Error("missing " + entityName + " id")
		WriteBadRequest(w, "missing_"+entityName+"_id", entityName+" id must be set", nil)
		return "", false
	}
	return id, true
}
