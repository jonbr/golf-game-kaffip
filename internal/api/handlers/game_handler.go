package handlers

import (
	"encoding/json"
	"errors"
	"golf-game-kaffip/internal/api"
	"golf-game-kaffip/internal/api/dto"
	domainCourse "golf-game-kaffip/internal/domain/course"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) CreateGame(w http.ResponseWriter, r *http.Request) {
	ctx, logger := startRequest(r, "create game")

	// 1. Bind JSON
	var req dto.CreateGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("invalid JSON", "error", err)
		api.WriteBadRequest(w, "invalid_input", "invalid JSON payload", nil)
		return
	}

	// 2. Execute service
	game, err := h.GameService.CreateGame(ctx, req)
	if err != nil {
		if errors.Is(err, domainCourse.ErrCourseNotFound) {
			logger.Info("create game failed: course not found", "course_id", req.CourseID)
			api.WriteNotFound(w, "course_not_found", "course does not exist", nil)
			return
		}
		logger.Error("create game failed", "error", err)
		api.WriteError(w, err)
		return
	}

	// 3. Send response
	api.JSON(w, http.StatusCreated, dto.CreateGameResponse{
		GameID: game.ID,
	})
}

func (h *Handler) GetGames(w http.ResponseWriter, r *http.Request) {
	ctx, logger := startRequest(r, "get games")

	status := r.URL.Query().Get("status") // "active", "finished", "all", ""
	logger.Debug("get games", "status", status)

	// 2. Excecute service
	game, err := h.GameService.GetGames(ctx, status)
	if err != nil {
		logger.Error("get games failed", "error", err)
		api.WriteError(w, err)
		return
	}

	// 3. Success
	api.JSON(w, http.StatusOK, game)
}

func (h *Handler) GetGame(w http.ResponseWriter, r *http.Request) {
	ctx, logger := startRequest(r, "get game")

	gameID, ok := parseGameID(w, r, logger)
	if !ok {
		return
	}

	// 2. Excecute service
	game, err := h.GameService.GetGame(ctx, gameID)
	if err != nil {
		logger.Error("get game failed", "game_id", gameID, "error", err)
		api.WriteError(w, err)
		return
	}

	// 3. Success
	api.JSON(w, http.StatusOK, game)
}

func (h *Handler) SetHoleScore(w http.ResponseWriter, r *http.Request) {
	ctx, logger := startRequest(r, "set hole score")

	// parse query parameters
	gameID, ok := parseGameID(w, r, logger)
	if !ok {
		return
	}
	holeNumber, ok := parseHoleNumber(w, r, logger)
	if !ok {
		return
	}

	var req dto.SetHoleScoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("invalid JSON", "error", err)
		api.WriteBadRequest(w, "invalid_input", "invalid JSON payload", nil)
		return
	}

	game, err := h.GameService.SetHoleScore(ctx, gameID, holeNumber, req.Scores)
	if err != nil {
		logger.Error("set hole score failed", "game_id", gameID, "hole_number", holeNumber, "error", err)
		api.WriteError(w, err)
		return
	}

	logger.Info("hole score set", "game_id", gameID, "hole_number", holeNumber)
	api.JSON(w, http.StatusOK, game)
}

func (h *Handler) FinishGame(w http.ResponseWriter, r *http.Request) {
	ctx, logger := startRequest(r, "finish game")

	gameID, ok := parseGameID(w, r, logger)
	if !ok {
		return
	}

	if err := h.GameService.FinishGame(ctx, gameID); err != nil {
		logger.Error("finish game failed", "game_id", gameID, "error", err)
		api.WriteError(w, err)
		return
	}

	// 3. Success
	logger.Info("game finished", "game_id", gameID)
	w.WriteHeader(http.StatusNoContent)
}

func parseGameID(w http.ResponseWriter, r *http.Request, logger *slog.Logger) (string, bool) {
	id := chi.URLParam(r, "id")
	if id == "" {
		logger.Error("missing game id")
		api.WriteBadRequest(w, "missing_game_id", "game id must be set", nil)
		return "", false
	}
	return id, true
}

// parseHoleNumber extracts and validates the holeNumber path parameter.
// On failure it writes a 400 response itself and returns ok=false.
func parseHoleNumber(w http.ResponseWriter, r *http.Request, logger *slog.Logger) (int, bool) {
	holeNumberStr := chi.URLParam(r, "holeNumber")
	holeNumber, err := strconv.Atoi(holeNumberStr)
	if err != nil {
		logger.Error("invalid hole number", "hole_number", holeNumberStr)
		api.WriteBadRequest(w, "invalid_hole_number", "hole number must be an integer", nil)
		return 0, false
	}
	return holeNumber, true
}
