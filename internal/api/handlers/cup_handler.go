package handlers

import (
	"encoding/json"
	"golf-game-kaffip/internal/api"
	"golf-game-kaffip/internal/api/dto"
	"golf-game-kaffip/internal/domain/cup"
	"net/http"
)

func (h *Handler) CreateMatchPlayCup(w http.ResponseWriter, r *http.Request) {
	ctx, logger := startRequest(r, "create match play cup")

	var req dto.CreateMatchPlayCupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("invalid JSON", "error", err)
		api.WriteBadRequest(w, "invalid_input", "invalid JSON payload", nil)
		return
	}

	c, err := h.MatchPlayCupService.CreateCup(ctx, req)
	if err != nil {
		logger.Error("create wolf cup failed", "error", err)
		api.WriteError(w, err)
		return
	}

	api.JSON(w, http.StatusCreated, mapCupToResponse(c, cup.Score{}))
}

func (h *Handler) GetCup(w http.ResponseWriter, r *http.Request) {
	ctx, logger := startRequest(r, "get cup")

	cupID, ok := parseID(w, r, logger, "cup")
	if !ok {
		return
	}

	detail, err := h.CupService.GetCup(ctx, cupID)
	if err != nil {
		logger.Error("get cup failed", "cup_id", cupID, "error", err)
		api.WriteError(w, err)
		return
	}

	api.JSON(w, http.StatusOK, mapCupToResponse(detail.Cup, detail.Score))

}

func (h *Handler) FinishCup(w http.ResponseWriter, r *http.Request) {
	ctx, logger := startRequest(r, "finish cup")

	cupID, ok := parseID(w, r, logger, "cup")
	if !ok {
		return
	}

	if err := h.CupService.FinishCup(ctx, cupID); err != nil {
		logger.Error("finish cup failed", "cup_id", cupID, "error", err)
		api.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func mapCupToResponse(c *cup.Cup, score cup.Score) dto.CupResponse {
	roster := make([]dto.CupRosterEntry, 0, len(c.Players))
	for playerID, side := range c.Players {
		roster = append(roster, dto.CupRosterEntry{
			PlayerID: playerID,
			Side:     string(side),
		})
	}

	return dto.CupResponse{
		ID:       c.ID,
		Name:     c.Name,
		Roster:   roster,
		MatchIDs: c.MatchIDs,
		Score: dto.CupScoreResponse{
			TeamA: score.TeamA,
			TeamB: score.TeamB,
		},
		FinishedAt: c.FinishedAt,
	}
}
