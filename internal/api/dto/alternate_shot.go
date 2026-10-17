package dto

type CreateAlternateShotRequest struct {
	CourseID string  `json:"course_id"`
	Variant  string  `json:"variant"` // "gross" or "net"
	TeamA    []int64 `json:"team_a"`  // exactly 2 players
	TeamB    []int64 `json:"team_b"`  // exactly 2 players
}

// SetAlternateShotHoleScoreRequest carries one gross score per team,
// since only one ball is in play per side.
type SetAlternateShotHoleScoreRequest struct {
	TeamAGross int `json:"team_a_gross"`
	TeamBGross int `json:"team_b_gross"`
}

type AlternateShotHoleResultResponse struct {
	Hole       HoleInfoResponse `json:"hole"`
	TeamAGross int              `json:"team_a_gross"`
	TeamBGross int              `json:"team_b_gross"`
	TeamANet   int              `json:"team_a_net"`
	TeamBNet   int              `json:"team_b_net"`
	WinnerTeam string           `json:"winner_team,omitempty"` // "" if halved
}

type AlternateShotResponse struct {
	GameHeaderResponse
	TeamA       []PlayerRoleResponse                       `json:"team_a"`
	TeamB       []PlayerRoleResponse                       `json:"team_b"`
	Status      MatchPlayStatusResponse                    `json:"status"`
	HoleResults map[string]AlternateShotHoleResultResponse `json:"hole_results"`
}
