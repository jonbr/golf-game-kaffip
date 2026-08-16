package handlers

import (
	"golf-game-kaffip/internal/api"
	"net/http"
)

func (h *Handler) SearchCourses(w http.ResponseWriter, r *http.Request) {
	ctx, logger := startRequest(r, "search courses")

	query := r.URL.Query().Get("q")
	if len(query) < 3 {
		api.WriteBadRequest(w, "query_too_short", "query must be at least 3 characters", nil)
		return
	}

	results, err := h.GameService.SearchCourses(ctx, query)
	if err != nil {
		logger.Error("course search failed", "query", query, "error", err)
		api.WriteError(w, err)
		return
	}

	api.JSON(w, http.StatusOK, results)
}
