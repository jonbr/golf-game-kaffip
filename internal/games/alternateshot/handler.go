package alternateshot

import (
	"encoding/json"
	"errors"
	"golf-game-kaffip/internal/api"
	"golf-game-kaffip/internal/api/dto"
	"net/http"
	"strconv"

	domainCourse "golf-game-kaffip/internal/domain/course"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx, logger := api.StartRequest(r, "create alternate_shot game")

	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("invalid JSON", "error", err)
		api.WriteBadRequest(w, "invalid_input", "invalid JSON payload", nil)
		return
	}

	gameID, err := h.service.CreateGame(ctx, req)
	if err != nil {
		if errors.Is(err, domainCourse.ErrCourseNotFound) {
			logger.Info("create alternate_shot game failed: course not found", "course_id", req.CourseID)
			api.WriteNotFound(w, "course_not_found", "course does not exist", nil)
			return
		}
		logger.Error("create alternate_shot game failed", "path", r.URL.Path, "error", err)
		api.WriteError(w, err)
		return
	}

	api.JSON(w, http.StatusCreated, dto.CreateGameResponse{GameID: gameID})
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	ctx, logger := api.StartRequest(r, "get alternate_shot game")

	id, ok := api.ParseID(w, r, logger, "game")
	if !ok {
		return
	}

	g, err := h.service.GetGame(ctx, id)
	if err != nil {
		logger.Error("get game alternate_shot failed", "game_id", id, "error", err)
		api.WriteError(w, err)
		return
	}

	api.JSON(w, http.StatusOK, NewResponse(g))
}

func (h *Handler) SetHoleScore(w http.ResponseWriter, r *http.Request) {
	ctx, logger := api.StartRequest(r, "set alternate_shot hole score")

	id, ok := api.ParseID(w, r, logger, "game")
	if !ok {
		return
	}

	holeNumber, err := strconv.Atoi(chi.URLParam(r, "holeNumber"))
	if err != nil {
		api.WriteBadRequest(w, "invalid_hole_number", "hole number must be an integer", nil)
		return
	}

	var req SetHoleScoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("invalid JSON", "error", err)
		api.WriteBadRequest(w, "invalid_input", "invalid JSON payload", nil)
		return
	}

	g, err := h.service.SetHoleScore(ctx, id, holeNumber, req)
	if err != nil {
		logger.Error("set hole score alternate_shot failed", "game_id", id, "hole", holeNumber, "error", err)
		api.WriteError(w, err)
		return
	}

	api.JSON(w, http.StatusOK, NewResponse(g))
}
