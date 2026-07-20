package handlers

import (
	"encoding/json"
	"golf-game-kaffip/internal/api"
	"golf-game-kaffip/internal/api/dto"
	"golf-game-kaffip/internal/logging"
	"net/http"
	"strconv"

	playerDomain "golf-game-kaffip/internal/domain/player"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) CreatePlayer(w http.ResponseWriter, r *http.Request) {
	ctx, logger := startRequest(r, "create player")

	// 1. Bind JSON
	var req dto.CreatePlayerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("invalid JSON", "error", err)
		api.WriteBadRequest(w, "invalid_input", "invalid JSON payload", nil)
		return
	}

	// 2. Execute service
	p, err := h.PlayerService.CreatePlayer(ctx, req.Name, req.Email, req.Handicap)
	if err != nil {
		logger.Error("create player failed", "error", err)
		api.WriteError(w, err)
		return
	}

	// 3. Build response DTO and send
	api.JSON(w, http.StatusCreated, dto.CreatePlayerResponse{
		ID:       p.ID,
		Name:     p.Name,
		Email:    p.Email,
		Handicap: p.Handicap,
	})
}

func (h *Handler) GetPlayers(w http.ResponseWriter, r *http.Request) {
	ctx, logger := startRequest(r, "get players")

	// 1. Execute service
	players, err := h.PlayerService.GetPlayers(ctx)
	if err != nil {
		logger.Error("get players failed", "error", err)
		api.WriteError(w, err)
		return
	}

	// 2. Build response DTO slice
	resp := make([]dto.GetPlayerResponse, len(players))
	for i, p := range players {
		resp[i] = toPlayerResponse(p)
	}

	// 3. Send response
	api.JSON(w, http.StatusOK, resp)
}

func (h *Handler) GetPlayer(w http.ResponseWriter, r *http.Request) {
	ctx, logger := startRequest(r, "get player")

	// 1. Parse and validate URL parameter
	id, ok := parseIDParam(w, r, "id")
	if !ok {
		return
	}

	// 2. Execute service
	p, err := h.PlayerService.GetPlayer(ctx, id)
	if err != nil {
		logger.Error("get player failed", "error", err)
		api.WriteError(w, err)
		return
	}

	// 3. Build response DTO and send
	api.JSON(w, http.StatusOK, toPlayerResponse(p))
}

func (h *Handler) UpdatePlayer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logging.FromCtx(ctx).Info("update player")

	// 1. Parse and validate URL parameter
	id, ok := parseIDParam(w, r, "id")
	if !ok {
		return
	}

	// 2. Bind JSON
	var req dto.UpdatePlayerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteBadRequest(w, "invalid_input", "invalid JSON payload", nil)
		return
	}

	// 2. Execute service
	updated, err := h.PlayerService.UpdatePlayer(ctx, id, playerDomain.UpdatePlayerParams{
		Name:     req.Name,
		Handicap: req.Handicap,
	})
	if err != nil {
		api.WriteError(w, err)
		return
	}

	// 4. Success
	api.JSON(w, http.StatusOK, toPlayerResponse(updated))
}

func (h *Handler) DeletePlayer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, ok := parseIDParam(w, r, "id")
	if !ok {
		return
	}

	logging.FromCtx(ctx).Info("delete player", "player_id", id)

	// 2. Execute service
	if err := h.PlayerService.SoftDeletePlayer(ctx, id); err != nil {
		api.WriteError(w, err)
		return
	}

	// 3. Success (204 No Content)
	api.JSON(w, http.StatusNoContent, nil)
}

// parseIDParam extracts and validates an int64 path parameter.
// On failure it writes a 400 response itself and returns ok=false.
func parseIDParam(w http.ResponseWriter, r *http.Request, param string) (id int64, ok bool) {
	idstr := chi.URLParam(r, param)
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		api.WriteBadRequest(w, "invalid_"+param, param+" must be a number", nil)
		return 0, false
	}
	return id, true
}

func toPlayerResponse(p *playerDomain.Player) dto.GetPlayerResponse {
	return dto.GetPlayerResponse{
		ID:       p.ID,
		Name:     p.Name,
		Handicap: p.Handicap,
	}
}
