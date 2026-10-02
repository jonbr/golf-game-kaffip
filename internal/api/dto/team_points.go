package dto

import "time"

type CreateTeamPointsRequest struct {
	CourseID string  `json:"course_id"`
	Variant  string  `json:"variant"` // "gross" or "net"
	TeamA    []int64 `json:"team_a"`
	TeamB    []int64 `json:"team_b"`
}

type TeamPointsResponse struct {
	ID           string                        `json:"id"`
	GameType     string                        `json:"game_type"`
	Variant      string                        `json:"variant"`
	Course       CourseSummaryResponse         `json:"course"`
	TeamA        []PlayerRoleResponse          `json:"team_a"`
	TeamB        []PlayerRoleResponse          `json:"team_b"`
	CurrentHole  int                           `json:"current_hole"`
	StartingLead int                           `json:"starting_lead"`
	MatchScore   MatchScoreResponse            `json:"match_score"`
	HoleResults  map[string]HoleResultResponse `json:"hole_results"`
	FinishedAt   *time.Time                    `json:"finished_at"`
}
