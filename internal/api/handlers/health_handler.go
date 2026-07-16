package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	status := "ok"
	message := "API is healthy"
	statusCode := http.StatusOK
	dbStatus := "ok"

	if err := h.DB.Ping(pingCtx); err != nil {
		status = "degraded"
		message = "API is running but the database is unreachable"
		statusCode = http.StatusServiceUnavailable
		dbStatus = "unreachable"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  status,
		"service": "golf_game_proto",
		"message": message,
		"db":      dbStatus,
	})
}
