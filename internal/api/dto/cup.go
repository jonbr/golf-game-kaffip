package dto

import "time"

type CupRosterEntry struct {
	PlayerID int64  `json:"player_id"`
	Side     string `json:"side"` // "A" or "B"
}

// CupBaseRequest holds the fields every Cup create request shares. It is
// embedded in the per-format requests below, so the JSON payload stays flat.
type CupBaseRequest struct {
	Name   string           `json:"name"`
	Roster []CupRosterEntry `json:"roster"`
}

type CreateMatchPlayCupRequest struct {
	CupBaseRequest
	Matches []CreateMatchPlayRequest `json:"matches"`
}

type CreateTeamPointsCupRequest struct {
	CupBaseRequest
	Matches []CreateTeamPointsRequest `json:"matches"`
}

type CreateWolfCupRequest struct {
	CupBaseRequest
	Matches []CreateWolfGameRequest `json:"matches"`
}

type CupResponse struct {
	ID         string           `json:"id"`
	Name       string           `json:"name"`
	Roster     []CupRosterEntry `json:"roster"`
	MatchIDs   []string         `json:"match_ids"`
	Score      CupScoreResponse `json:"score"`
	FinishedAt *time.Time       `json:"finished_at"`
}

type CupScoreResponse struct {
	TeamA float64 `json:"team_a"`
	TeamB float64 `json:"team_b"`
}
