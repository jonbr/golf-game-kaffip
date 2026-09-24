package handlers

import (
	"encoding/json"
	"golf-game-kaffip/internal/api"
	"golf-game-kaffip/internal/api/dto"
	domainGame "golf-game-kaffip/internal/domain/game"
	"golf-game-kaffip/internal/domain/player"

	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) GetGames(w http.ResponseWriter, r *http.Request) {
	ctx, logger := startRequest(r, "get games")

	status := r.URL.Query().Get("status") // "active", "finished", "all", ""
	logger.Debug("get games", "status", status)

	// 2. Excecute service
	summaries, err := h.GameService.GetGames(ctx, status)
	if err != nil {
		logger.Error("get games failed", "error", err)
		api.WriteError(w, err)
		return
	}

	// 3. Parse response
	resp := make([]dto.GameSummaryResponse, len(summaries))
	for i, g := range summaries {
		resp[i] = mapGameSummaryToResponse(g)
	}

	// 4. Success
	api.JSON(w, http.StatusOK, resp)
}

func (h *Handler) SetHoleScore(w http.ResponseWriter, r *http.Request) {
	ctx, logger := startRequest(r, "set hole score")

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

	gameType, err := h.GameService.GetGameType(ctx, gameID)
	if err != nil {
		logger.Error("failed to resolve game type", "game_id", gameID, "error", err)
		api.WriteError(w, err)
		return
	}

	// respond handles the shared error-logging/response shape so each
	// case below only supplies what's actually different: which service
	// to call and how to map its result.
	respond := func(err error, mapResponse func() any) {
		if err != nil {
			logger.Error("set hole score failed", "game_id", gameID, "hole_number", holeNumber, "error", err)
			api.WriteError(w, err)
			return
		}
		api.JSON(w, http.StatusOK, mapResponse())
	}

	switch gameType {
	case domainGame.GameTypeMatchPlay:
		game, err := h.MatchPlayService.SetHoleScore(ctx, gameID, holeNumber, req)
		respond(err, func() any { return mapGameToResponse(game) })

	case domainGame.GameTypeTeamPoints:
		game, err := h.TeamPointsService.SetHoleScore(ctx, gameID, holeNumber, req)
		respond(err, func() any { return mapGameToResponse(game) })

	case domainGame.GameTypeWolf:
		wolfReq := dto.SetWolfHoleScoreRequest{
			WolfPlayerID: *req.WolfPlayerID,
			Mode:         *req.Mode,
			PartnerID:    req.PartnerID,
			Scores:       req.Scores,
		}
		g, err := h.WolfGameService.SetHoleScore(ctx, gameID, holeNumber, wolfReq)
		respond(err, func() any { return mapWolfGameToResponse(g) })

	default:
		api.WriteBadRequest(w, "unsupported_game_type", "this game type does not support hole scoring", nil)
	}
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

func mapTeamPlayersToRoles(teamA, teamB []*player.Player) []dto.PlayerRoleResponse {
	teamLabelA := "A"
	teamLabelB := "B"

	roles := make([]dto.PlayerRoleResponse, 0, len(teamA)+len(teamB))
	for _, p := range teamA {
		roles = append(roles, dto.PlayerRoleResponse{
			PlayerID: p.ID, Team: &teamLabelA, Name: p.Name, Email: p.Email, Handicap: p.Handicap,
		})
	}
	for _, p := range teamB {
		roles = append(roles, dto.PlayerRoleResponse{
			PlayerID: p.ID, Team: &teamLabelB, Name: p.Name, Email: p.Email, Handicap: p.Handicap,
		})
	}
	return roles
}

func mapGameToResponse(g *domainGame.Game) dto.GameResponse {
	holeResultsResp := make(map[string]dto.HoleResultResponse, len(g.HoleResults))
	for holeNum, hr := range g.HoleResults {
		scores := make([]dto.PlayerScoreResponse, 0, len(hr.Scores))
		for _, s := range hr.Scores {
			scores = append(scores, dto.PlayerScoreResponse{
				PlayerID: s.PlayerID,
				Gross:    s.Gross,
				Net:      s.Net,
			})
		}

		grossBonuses := make([]dto.GrossBonusResponse, 0, len(hr.GrossBonuses))
		for _, gb := range hr.GrossBonuses {
			grossBonuses = append(grossBonuses, dto.GrossBonusResponse{
				PlayerID: gb.PlayerID,
				Bonus:    gb.Bonus,
			})
		}

		holeResultsResp[strconv.Itoa(holeNum)] = dto.HoleResultResponse{
			LowScoreWinnerTeam:  hr.LowScoreWinnerTeam,
			TeamTotalWinnerTeam: hr.TeamTotalWinnerTeam,
			Scores:              scores,
			GrossBonuses:        grossBonuses,
		}
	}

	return dto.GameResponse{
		ID:           g.ID,
		GameType:     string(g.GameType),
		Variant:      string(g.Variant),
		Course:       dto.CourseSummaryResponse{ID: g.Course.ID, Name: g.Course.Name},
		Players:      mapTeamPlayersToRoles(g.TeamA, g.TeamB),
		CurrentHole:  g.CurrentHole,
		StartingLead: g.StartingLead,
		MatchScore:   dto.MatchScoreResponse{TeamA: g.MatchScore.TeamA, TeamB: g.MatchScore.TeamB},
		HoleResults:  holeResultsResp,
		FinishedAt:   g.FinishedAt,
	}
}

func mapGameSummaryToResponse(g *domainGame.GameSummary) dto.GameSummaryResponse {
	return dto.GameSummaryResponse{
		ID:          g.ID,
		GameType:    string(g.GameType),
		Course:      dto.CourseSummaryResponse{ID: g.CourseID, Name: g.CourseName},
		CurrentHole: g.CurrentHole,
		TotalHoles:  g.TotalHoles,
		FinishedAt:  g.FinishedAt,
	}
}
