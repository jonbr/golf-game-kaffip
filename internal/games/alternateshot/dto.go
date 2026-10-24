package alternateshot

import (
	"golf-game-kaffip/internal/api/dto"
	"golf-game-kaffip/internal/domain/player"
	"strconv"
)

// CreateRequest starts an alternate-shot (Greensome/Foursomes) match: two
// teams of exactly two players, match-play scoring.
type CreateRequest struct {
	CourseID string  `json:"course_id"`
	Variant  string  `json:"variant"`
	TeamA    []int64 `json:"team_a"`
	TeamB    []int64 `json:"team_b"`
}

// SetHoleScoreRequest carries one gross score per team for a hole.
// since only one ball is in play per side.
type SetHoleScoreRequest struct {
	TeamAGross int `json:"team_a_gross"`
	TeamBGross int `json:"team_b_gross"`
}

type HoleResultResponse struct {
	Hole       int    `json:"hole"`
	TeamAGross int    `json:"team_a_gross"`
	TeamBGross int    `json:"team_b_gross"`
	TeamANet   int    `json:"team_a_net"`
	TeamBNet   int    `json:"team_b_net"`
	WinnerTeam string `json:"winner_team,omitempty"` // "" if halved
}

type Response struct {
	dto.GameHeaderResponse
	TeamA       []dto.PlayerRoleResponse      `json:"team_a"`
	TeamB       []dto.PlayerRoleResponse      `json:"team_b"`
	Status      dto.MatchPlayStatusResponse   `json:"status"`
	HoleResults map[string]HoleResultResponse `json:"hole_results"`
}

// NewResponse builds the API view of a game.
func NewResponse(g *Game) Response {
	st := g.Status()

	results := make(map[string]HoleResultResponse, len(g.HoleResults))
	for n, r := range g.HoleResults {
		results[strconv.Itoa(n)] = HoleResultResponse{
			Hole:       n,
			TeamAGross: r.TeamAGross,
			TeamBGross: r.TeamBGross,
			TeamANet:   r.TeamANet,
			TeamBNet:   r.TeamBNet,
			WinnerTeam: r.WinnerTeam(),
		}
	}

	return Response{
		GameHeaderResponse: dto.GameHeaderResponse{
			ID:          g.ID,
			GameType:    string(GameType),
			Variant:     string(g.Variant),
			Course:      dto.CourseSummaryResponse{ID: g.Course.ID, Name: g.Course.Name},
			CurrentHole: g.CurrentHole,
			FinishedAt:  g.FinishedAt,
		},
		TeamA:       roles(g.TeamA),
		TeamB:       roles(g.TeamB),
		Status:      dto.MatchPlayStatusResponse{Display: st.Display, WinnerTeam: st.WinnerTeam, Closed: st.Closed},
		HoleResults: results,
	}
}

func roles(team []*player.Player) []dto.PlayerRoleResponse {
	out := make([]dto.PlayerRoleResponse, 0, len(team))
	for i, p := range team {
		seat := i
		out = append(out, dto.PlayerRoleResponse{
			PlayerID: p.ID,
			Seat:     &seat,
			Name:     p.Name,
			Handicap: p.Handicap,
		})
	}
	return out
}
