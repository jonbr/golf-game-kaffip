package dto

type CreateTeamPointsRequest struct {
	CourseID string  `json:"course_id"`
	Variant  string  `json:"variant"` // "gross" or "net"
	TeamA    []int64 `json:"team_a"`
	TeamB    []int64 `json:"team_b"`
}

type TeamPointsResponse struct {
	GameHeaderResponse
	TeamA        []PlayerRoleResponse          `json:"team_a"`
	TeamB        []PlayerRoleResponse          `json:"team_b"`
	StartingLead int                           `json:"starting_lead"`
	MatchScore   MatchScoreResponse            `json:"match_score"`
	HoleResults  map[string]HoleResultResponse `json:"hole_results"`
}
