package handlers

import (
	"encoding/json"
	"errors"
	"golf-game-kaffip/internal/api"
	"golf-game-kaffip/internal/api/dto"
	domainCourse "golf-game-kaffip/internal/domain/course"
	domainGame "golf-game-kaffip/internal/domain/game"
	"net/http"
	"strconv"
)

func (h *Handler) CreateTeamPoints(w http.ResponseWriter, r *http.Request) {
	ctx, logger := startRequest(r, "create team points game")

	// 1. Bind JSON
	var req dto.CreateTeamPointsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("invalid JSON", "error", err)
		api.WriteBadRequest(w, "invalid_input", "invalid JSON payload", nil)
		return
	}

	// 2. Execute service
	game, err := h.TeamPointsService.CreateGame(ctx, domainGame.GameTypeTeamPoints, req)
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

func (h *Handler) GetTeamPointsGame(w http.ResponseWriter, r *http.Request) {
	ctx, logger := startRequest(r, "get game")

	id, ok := parseGameID(w, r, logger)
	if !ok {
		return
	}

	game, err := h.GameService.GetGame(ctx, id)
	if err != nil {
		logger.Error("get team_points game failed", "game_id", id, "error", err)
		api.WriteError(w, err)
		return
	}

	if game.GameType != domainGame.GameTypeTeamPoints {
		api.WriteBadRequest(w, "wrong_game_type", "this game is not a team_points game", map[string]any{
			"game_id": game.ID, "actual_type": string(game.GameType),
		})
		return
	}

	api.JSON(w, http.StatusOK, mapTeamPointsToResponse(game))
}

func mapTeamPointsToResponse(g *domainGame.Game) dto.TeamPointsResponse {
	holeResultsResp := make(map[string]dto.HoleResultResponse, len(g.HoleResults))
	for holeNum, hr := range g.HoleResults {
		scores := make([]dto.PlayerScoreResponse, 0, len(hr.Scores))
		for _, s := range hr.Scores {
			scores = append(scores, dto.PlayerScoreResponse{
				PlayerID: s.PlayerID, Gross: s.Gross, Net: s.Net,
			})
		}

		grossBonuses := make([]dto.GrossBonusResponse, 0, len(hr.GrossBonuses))
		for _, gb := range hr.GrossBonuses {
			grossBonuses = append(grossBonuses, dto.GrossBonusResponse{
				PlayerID: gb.PlayerID, TeamID: gb.TeamID, Bonus: gb.Bonus,
			})
		}

		holeResultsResp[strconv.Itoa(holeNum)] = dto.HoleResultResponse{
			LowScoreWinnerTeam:  hr.LowScoreWinnerTeam,
			TeamTotalWinnerTeam: hr.TeamTotalWinnerTeam,
			Scores:              scores,
			GrossBonuses:        grossBonuses,
		}
	}

	return dto.TeamPointsResponse{
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
