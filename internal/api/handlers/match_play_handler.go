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
	game, err := h.MatchPlayService.CreateGame(ctx, req)
	if err != nil {
		if errors.Is(err, domainCourse.ErrCourseNotFound) {
			logger.Info("create game match_play failed: course not found", "course_id", req.CourseID)
			api.WriteNotFound(w, "course_not_found", "course does not exist", nil)
			return
		}
		logger.Error("create game match_play failed", "path", r.URL.Path, "error", err)
		api.WriteError(w, err)
		return
	}

	// 3. Send response
	api.JSON(w, http.StatusCreated, dto.CreateGameResponse{
		GameID: game.ID,
	})
}

func (h *Handler) GetMatchPlayGame(w http.ResponseWriter, r *http.Request) {
	ctx, logger := startRequest(r, "get game")

	id, ok := parseGameID(w, r, logger)
	if !ok {
		return
	}

	game, err := h.GameService.GetGame(ctx, id)
	if err != nil {
		logger.Error("get match_play failed", "game_id", id, "error", err)
		api.WriteError(w, err)
		return
	}

	if game.GameType != domainGame.GameTypeMatchPlay {
		api.WriteBadRequest(w, "wrong_game_type", "this game is not a match_play game", map[string]any{
			"game_id": game.ID, "actual_type": string(game.GameType),
		})
		return
	}

	api.JSON(w, http.StatusOK, mapMatchPlayToResponse(game))
}

func mapMatchPlayToResponse(g *domainGame.Game) dto.MatchPlayResponse {
	status := g.MatchPlayStatus()

	holeResultsResp := make(map[string]dto.HoleResultResponse, len(g.HoleResults))
	for holeNum, hr := range g.HoleResults {
		scores := make([]dto.PlayerScoreResponse, 0, len(hr.Scores))
		for _, s := range hr.Scores {
			scores = append(scores, dto.PlayerScoreResponse{
				PlayerID: s.PlayerID, Gross: s.Gross, Net: s.Net,
			})
		}
		holeResultsResp[strconv.Itoa(holeNum)] = dto.HoleResultResponse{
			LowScoreWinnerTeam: hr.LowScoreWinnerTeam,
			Scores:             scores,
		}
	}

	playerA := g.TeamA[0]
	playerB := g.TeamB[0]
	teamLabelA := "A"
	teamLabelB := "B"

	return dto.MatchPlayResponse{
		ID:       g.ID,
		GameType: string(g.GameType),
		Variant:  string(g.Variant),
		Course:   dto.CourseSummaryResponse{ID: g.Course.ID, Name: g.Course.Name},
		PlayerA: dto.PlayerRoleResponse{
			PlayerID: playerA.ID, Team: &teamLabelA, Name: playerA.Name, Email: playerA.Email, Handicap: playerA.Handicap,
		},
		PlayerB: dto.PlayerRoleResponse{
			PlayerID: playerB.ID, Team: &teamLabelB, Name: playerB.Name, Email: playerB.Email, Handicap: playerB.Handicap,
		},
		CurrentHole: g.CurrentHole,
		Status: dto.MatchPlayStatusResponse{
			Display: status.Display, WinnerTeam: status.WinnerTeam, Closed: status.Closed,
		},
		HoleResults: holeResultsResp,
		FinishedAt:  g.FinishedAt,
	}
}
