package handlers

import (
	"encoding/json"
	"golf-game-kaffip/internal/api"
	"golf-game-kaffip/internal/api/dto"
	"net/http"
)

// TODO: Implement alternate shot game handler
func (h *Handler) CreateAlternateShot(w http.ResponseWriter, r *http.Request) {
	//ctx, logger := startRequest(r, "create alternate shot game")

	// 1. Bind JSON
	var req dto.CreateAlternateShotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		//logger.Error("invalid JSON", "error", err)
		api.WriteBadRequest(w, "invalid_input", "invalid JSON payload", nil)
		return
	}

	// 2. Execute service
	/*game, err := h.AlternateShotService.CreateGame(ctx, req)
	if err != nil {
		if errors.Is(err, domainCourse.ErrCourseNotFound) {
			logger.Info("create game alternate_shot failed: course not found", "course_id", req.CourseID)
			api.WriteNotFound(w, "course_not_found", "course does not exist", nil)
			return
		}
		logger.Error("create game alternate_shot failed", "path", r.URL.Path, "error", err)
		api.WriteError(w, err)
		return
	}*/

	// 3. Send response
	api.JSON(w, http.StatusCreated, dto.CreateGameResponse{
		GameID: "nil", //game.ID,
	})

}
