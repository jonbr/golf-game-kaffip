package dto

import "time"

type CreateWolfGameRequest struct {
	CourseID  string  `json:"course_id"`
	PlayerIDs []int64 `json:"players"`
}

type CreateWolfGameResponse struct {
	GameID string `json:"game_id"`
}

type SetWolfHoleScoreRequest struct {
	WolfPlayerID int64              `json:"wolf_player_id"`
	Mode         string             `json:"mode"` // "partnered" or "lone"
	PartnerID    *int64             `json:"partner_id,omitempty"`
	Scores       []PlayerGrossScore `json:"scores"` // all 4 players
}

type WolfGameResponse struct {
	ID                  string                            `json:"id"`
	Course              CourseSummaryResponse             `json:"course"`
	Players             []PlayerRoleResponse              `json:"players"` // was []PlayerSummaryResponse
	CurrentHole         int                               `json:"current_hole"`
	CurrentWolfPlayerID int64                             `json:"current_wolf_player_id"`
	Standings           []WolfStandingResponse            `json:"standings"`
	HoleResults         map[string]WolfHoleResultResponse `json:"hole_results"`
	FinishedAt          *time.Time                        `json:"finished_at"`
}

type WolfStandingResponse struct {
	PlayerID int64 `json:"player_id"`
	Points   int   `json:"points"`
}

type WolfHoleResultResponse struct {
	WolfPlayerID int64                     `json:"wolf_player_id"`
	Mode         string                    `json:"mode"` // "partnered" or "lone"
	PartnerID    *int64                    `json:"partner_id,omitempty"`
	WinningSide  string                    `json:"winning_side"` // "wolf", "field", or "" if tied
	Scores       []WolfPlayerScoreResponse `json:"scores"`
}

type WolfPlayerScoreResponse struct {
	PlayerID int64 `json:"player_id"`
	Gross    int   `json:"gross"`
	Net      int   `json:"net"`
	Strokes  int   `json:"strokes"`
	Points   int   `json:"points"`
}
