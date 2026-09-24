package handlers

import (
	"encoding/json"
	"errors"
	"golf-game-kaffip/internal/api"
	"golf-game-kaffip/internal/api/dto"
	domainCourse "golf-game-kaffip/internal/domain/course"
	domainGame "golf-game-kaffip/internal/domain/game"
	"net/http"
)

func (h *Handler) CreateMatchPlay(w http.ResponseWriter, r *http.Request) {
	ctx, logger := startRequest(r, "create match play game")

	// 1. Bind JSON
	var req dto.CreateMatchPlayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("invalid JSON", "error", err)
		api.WriteBadRequest(w, "invalid_input", "invalid JSON payload", nil)
		return
	}

	// 2. Execute service
	game, err := h.MatchPlayService.CreateGame(ctx, domainGame.GameTypeMatchPlay, req)
	if err != nil {
		if errors.Is(err, domainCourse.ErrCourseNotFound) {
			logger.Info("create game failed: course not found", "course_id", req.CourseID)
			api.WriteNotFound(w, "course_not_found", "course does not exist", nil)
			return
		}
		logger.Error("create game failed", "path", r.URL.Path, "error", err)
		api.WriteError(w, err)
		return
	}

	// 3. Send response
	api.JSON(w, http.StatusCreated, dto.CreateGameResponse{
		GameID: game.ID,
	})
}
