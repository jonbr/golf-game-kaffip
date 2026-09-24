package dto

import "time"

type CreateGameResponse struct {
	GameID string `json:"game_id"`
}

type SetHoleScoreRequest struct {
	Scores []PlayerGrossScore `json:"scores"`

	// Wolf-only fields, ignored by match_play/team_points.
	WolfPlayerID *int64  `json:"wolf_player_id,omitempty"`
	Mode         *string `json:"mode,omitempty"`
	PartnerID    *int64  `json:"partner_id,omitempty"`
}

type PlayerGrossScore struct {
	PlayerID int64 `json:"player_id" binding:"required"`
	Gross    int   `json:"gross" binding:"required"`
}

type PlayerScoreResponse struct {
	PlayerID int64 `json:"player_id"`
	Gross    int   `json:"gross"`
	Net      int   `json:"net"`
}

type HoleResultResponse struct {
	Hole                HoleInfoResponse      `json:"hole"`
	Scores              []PlayerScoreResponse `json:"scores"`
	LowScoreWinnerTeam  string                `json:"low_score_winner_team,omitempty"`
	TeamTotalWinnerTeam string                `json:"team_total_winner_team,omitempty"`
	GrossBonuses        []GrossBonusResponse  `json:"gross_bonuses,omitempty"`
}

type HoleInfoResponse struct {
	Number      int `json:"number"`
	Par         int `json:"par"`
	StrokeIndex int `json:"stroke_index"`
}

type MatchScoreResponse struct {
	TeamA int `json:"team_a"`
	TeamB int `json:"team_b"`
}

type GrossBonusResponse struct {
	PlayerID int64  `json:"player_id"`
	TeamID   string `json:"team_id"`
	Bonus    int    `json:"bonus"`
}

type GameSummaryResponse struct {
	ID          string                `json:"id"`
	GameType    string                `json:"game_type"`
	Course      CourseSummaryResponse `json:"course"`
	CurrentHole int                   `json:"current_hole"`
	TotalHoles  int                   `json:"total_holes"`
	FinishedAt  *time.Time            `json:"finished_at"`
}
