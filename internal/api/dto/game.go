package dto

type CreateGameRequest struct {
	CourseID string  `json:"course_id"`
	TeamA    []int64 `json:"team_a"`
	TeamB    []int64 `json:"team_b"`
	Variant  string  `json:"variant"` // "gross" or "net"
}

type CreateGameResponse struct {
	GameID string `json:"game_id"`
}

type SubmitHoleScoreRequest struct {
	GameID string             `json:"-"` // injected from URL
	Hole   int                `json:"hole" binding:"required"`
	Scores []PlayerGrossScore `json:"scores" binding:"required,dive"`
}

type SetHoleScoreRequest struct {
	Scores []PlayerGrossScore `json:"scores"`
}

type PlayerGrossScore struct {
	PlayerID int64 `json:"player_id" binding:"required"`
	Gross    int   `json:"gross" binding:"required"`
}

type GameStateResponse struct {
	GameID      string              `json:"game_id"`
	Course      CourseResponse      `json:"course"`
	Teams       TeamsResponse       `json:"teams"`
	CurrentHole int                 `json:"current_hole"`
	HoleResult  *HoleResultResponse `json:"hole_result,omitempty"`
	MatchScore  MatchScoreResponse  `json:"match_score"`
}

type CourseResponse struct {
	ID    string         `json:"id"`
	Name  string         `json:"name"`
	Holes []HoleResponse `json:"holes"`
}

type HoleResponse struct {
	Number      int `json:"number"`
	Par         int `json:"par"`
	StrokeIndex int `json:"stroke_index"`
}

type TeamsResponse struct {
	TeamA []PlayerResponse `json:"team_a"`
	TeamB []PlayerResponse `json:"team_b"`
}

type PlayerResponse struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Handicap float64 `json:"handicap"`
}

type PlayerScoreResponse struct {
	PlayerID string `json:"player_id"`
	Gross    int    `json:"gross"`
	Net      int    `json:"net"`
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
	PlayerID string `json:"player_id"`
	TeamID   string `json:"team_id"`
	Bonus    int    `json:"bonus"`
}
