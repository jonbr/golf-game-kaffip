package dto

import "time"

type EventMatchInput struct {
	GameType string  `json:"game_type"` // "match_play" or "team_play"
	TeamA    []int64 `json:"team_a"`    // 1 player for match_play, 2 for team_play
	TeamB    []int64 `json:"team_b"`
}

type CreateEventRequest struct {
	CourseID string            `json:"course_id"`
	Variant  string            `json:"variant"`
	Matches  []EventMatchInput `json:"matches"`
}

type CreateEventResponse struct {
	EventID  string   `json:"event_id"`
	MatchIDs []string `json:"match_ids"`
}

type EventResponse struct {
	ID         string                `json:"id"`
	Course     CourseSummaryResponse `json:"course"`
	Variant    string                `json:"variant"`
	Score      EventScoreResponse    `json:"score"`
	Matches    []EventMatchResponse  `json:"matches"`
	FinishedAt *time.Time            `json:"finished_at"`
}

type CourseSummaryResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type EventScoreResponse struct {
	TeamA float64 `json:"team_a"`
	TeamB float64 `json:"team_b"`
}

type EventMatchResponse struct {
	GameID      string                  `json:"game_id"`
	GameType    string                  `json:"game_type"`
	TeamA       []PlayerSummaryResponse `json:"team_a"`
	TeamB       []PlayerSummaryResponse `json:"team_b"`
	CurrentHole int                     `json:"current_hole"`
	Status      MatchOutcomeResponse    `json:"status"`
}

type PlayerSummaryResponse struct {
	ID       int64   `json:"id"`
	Name     string  `json:"name"`
	Handicap float64 `json:"handicap"`
}

type MatchOutcomeResponse struct {
	Display    string `json:"display"`
	WinnerTeam string `json:"winner_team,omitempty"`
	Decided    bool   `json:"decided"`
}
