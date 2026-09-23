package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"path"
	"strconv"

	"golf-game-kaffip/internal/api"
	"golf-game-kaffip/internal/api/dto"
	domainCourse "golf-game-kaffip/internal/domain/course"
	"golf-game-kaffip/internal/domain/player"
	domainWolf "golf-game-kaffip/internal/domain/wolf"
)

func (h *Handler) CreateWolfGame(w http.ResponseWriter, r *http.Request) {
	ctx, logger := startRequest(r, "create "+path.Base(r.URL.Path)+" game")

	var req dto.CreateWolfGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("invalid JSON", "error", err)
		api.WriteBadRequest(w, "invalid_input", "invalid JSON payload", nil)
		return
	}

	game, err := h.WolfGameService.CreateGame(ctx, req)
	if err != nil {
		if errors.Is(err, domainCourse.ErrCourseNotFound) {
			logger.Info("create wolf game failed: course not found", "course_id", req.CourseID)
			api.WriteNotFound(w, "course_not_found", "course does not exist", nil)
			return
		}
		logger.Error("create wolf game failed", "path", r.URL.Path, "error", err)
		api.WriteError(w, err)
		return
	}

	api.JSON(w, http.StatusCreated, dto.CreateWolfGameResponse{
		GameID: game.ID,
	})
}

func (h *Handler) GetWolfGame(w http.ResponseWriter, r *http.Request) {
	ctx, logger := startRequest(r, "get wolf game")

	id, ok := parseGameID(w, r, logger)
	if !ok {
		return
	}

	game, err := h.WolfGameService.GetGame(ctx, id)
	if err != nil {
		logger.Error("get wolf game failed", "wolf_game_id", id, "error", err)
		api.WriteError(w, err)
		return
	}

	api.JSON(w, http.StatusOK, mapWolfGameToResponse(game))
}

func (h *Handler) SetWolfHoleScore(w http.ResponseWriter, r *http.Request) {}

func mapWolfPlayersToRoles(players [4]*player.Player) []dto.PlayerRoleResponse {
	roles := make([]dto.PlayerRoleResponse, 0, 4)
	for seat, p := range players {
		s := seat
		roles = append(roles, dto.PlayerRoleResponse{
			PlayerID: p.ID, Seat: &s, Name: p.Name, Email: p.Email, Handicap: p.Handicap,
		})
	}
	return roles
}

// mapWolfGameToResponse builds the API response for a Wolf game, including
// the unified players array, live standings, and whose turn it is to be
// Wolf — all derived fresh from HoleResults, never stored, so they can
// never drift out of sync with corrections.
func mapWolfGameToResponse(g *domainWolf.Game) dto.WolfGameResponse {
	standings := computeWolfStandings(g)
	currentWolfID := domainWolf.WolfForHole(g.Players, g.CurrentHole, standingsThroughHole(g, g.CurrentHole-1))

	standingsResp := make([]dto.WolfStandingResponse, 0, 4)
	for _, p := range g.Players {
		standingsResp = append(standingsResp, dto.WolfStandingResponse{
			PlayerID: p.ID,
			Points:   standings[p.ID],
		})
	}

	holeResultsResp := make(map[string]dto.WolfHoleResultResponse, len(g.HoleResults))
	for holeNum, hr := range g.HoleResults {
		scores := make([]dto.WolfPlayerScoreResponse, 0, len(hr.Scores))
		for _, s := range hr.Scores {
			scores = append(scores, dto.WolfPlayerScoreResponse{
				PlayerID: s.PlayerID,
				Gross:    s.Gross,
				Net:      s.Net,
				Strokes:  s.Strokes,
				Points:   hr.PointsAwarded[s.PlayerID],
			})
		}

		holeResultsResp[strconv.Itoa(holeNum)] = dto.WolfHoleResultResponse{
			WolfPlayerID: hr.WolfPlayerID,
			Mode:         string(hr.Mode),
			PartnerID:    hr.PartnerID,
			WinningSide:  string(hr.WinningSide),
			Scores:       scores,
		}
	}

	return dto.WolfGameResponse{
		ID:                  g.ID,
		Course:              dto.CourseSummaryResponse{ID: g.Course.ID, Name: g.Course.Name},
		Players:             mapWolfPlayersToRoles(g.Players),
		CurrentHole:         g.CurrentHole,
		CurrentWolfPlayerID: currentWolfID.ID,
		Standings:           standingsResp,
		HoleResults:         holeResultsResp,
		FinishedAt:          g.FinishedAt,
	}
}

// computeWolfStandings sums PointsAwarded across every recorded hole.
// Never stored — always recomputed, so corrections to any hole
// automatically ripple through correctly.
func computeWolfStandings(g *domainWolf.Game) map[int64]int {
	standings := make(map[int64]int, 4)
	for _, p := range g.Players {
		standings[p.ID] = 0
	}
	for _, hr := range g.HoleResults {
		for playerID, points := range hr.PointsAwarded {
			standings[playerID] += points
		}
	}
	return standings
}

// standingsThroughHole computes standings using only holes up to and
// including the given hole number — needed by WolfForHole, which must
// only see points earned before the hole it's determining Wolf for.
func standingsThroughHole(g *domainWolf.Game, throughHole int) map[int64]int {
	standings := make(map[int64]int, 4)
	for _, p := range g.Players {
		standings[p.ID] = 0
	}
	for holeNum, hr := range g.HoleResults {
		if holeNum > throughHole {
			continue
		}
		for playerID, points := range hr.PointsAwarded {
			standings[playerID] += points
		}
	}
	return standings
}
