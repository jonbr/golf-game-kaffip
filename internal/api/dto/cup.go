package dto

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
