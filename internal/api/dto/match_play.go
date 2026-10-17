package dto

type CreateMatchPlayRequest struct {
	CourseID string  `json:"course_id"`
	Variant  string  `json:"variant"` // "gross" or "net"
	TeamA    []int64 `json:"team_a"`
	TeamB    []int64 `json:"team_b"`
}

type MatchPlayResponse struct {
	GameHeaderResponse
	TeamA       []PlayerRoleResponse          `json:"team_a"`
	TeamB       []PlayerRoleResponse          `json:"team_b"`
	Status      MatchPlayStatusResponse       `json:"status"`
	HoleResults map[string]HoleResultResponse `json:"hole_results"`
}

type MatchPlayStatusResponse struct {
	Display    string `json:"display"`
	WinnerTeam string `json:"winner_team,omitempty"`
	Closed     bool   `json:"closed"`
}
