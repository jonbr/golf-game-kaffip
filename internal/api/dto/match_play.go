package dto

import "time"

type CreateMatchPlayRequest struct {
	CourseID string `json:"course_id"`
	Variant  string `json:"variant"`
	PlayerA  int64  `json:"player_a"`
	PlayerB  int64  `json:"player_b"`
}

type MatchPlayResponse struct {
	ID          string                        `json:"id"`
	GameType    string                        `json:"game_type"`
	Variant     string                        `json:"variant"`
	Course      CourseSummaryResponse         `json:"course"`
	PlayerA     PlayerRoleResponse            `json:"player_a"`
	PlayerB     PlayerRoleResponse            `json:"player_b"`
	CurrentHole int                           `json:"current_hole"`
	Status      MatchPlayStatusResponse       `json:"status"`
	HoleResults map[string]HoleResultResponse `json:"hole_results"`
	FinishedAt  *time.Time                    `json:"finished_at"`
}

type MatchPlayStatusResponse struct {
	Display    string `json:"display"`
	WinnerTeam string `json:"winner_team,omitempty"`
	Closed     bool   `json:"closed"`
}
