package dto

import "time"

type CreateTeamPointsRequest struct {
	CourseID string  `json:"course_id"`
	TeamA    []int64 `json:"team_a"`
	TeamB    []int64 `json:"team_b"`
	Variant  string  `json:"variant"` // "gross" or "net"
}

type TeamPointsResponse struct {
	ID           string                        `json:"id"`
	GameType     string                        `json:"game_type"`
	Variant      string                        `json:"variant"`
	Course       CourseSummaryResponse         `json:"course"`
	Players      []PlayerRoleResponse          `json:"players"`
	CurrentHole  int                           `json:"current_hole"`
	StartingLead int                           `json:"starting_lead"`
	MatchScore   MatchScoreResponse            `json:"match_score"`
	HoleResults  map[string]HoleResultResponse `json:"hole_results"`
	FinishedAt   *time.Time                    `json:"finished_at"`
}
