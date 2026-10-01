package dto

import "time"

type CupRosterEntry struct {
	PlayerID int64  `json:"player_id"`
	Side     string `json:"side"` // "A" or "B"
}

type CreateMatchPlayCupRequest struct {
	Name    string                   `json:"name"`
	Roster  []CupRosterEntry         `json:"roster"`
	Matches []CreateMatchPlayRequest `json:"matches"`
}

type CreateTeamPointsCupRequest struct {
	Name    string                    `json:"name"`
	Roster  []CupRosterEntry          `json:"roster"`
	Matches []CreateTeamPointsRequest `json:"matches"`
}

type CreateWolfCupRequest struct {
	Name    string                  `json:"name"`
	Roster  []CupRosterEntry        `json:"roster"`
	Matches []CreateWolfGameRequest `json:"matches"`
}

type CreateCupRequest struct {
	Name   string           `json:"name"`
	Roster []CupRosterEntry `json:"roster"`
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
