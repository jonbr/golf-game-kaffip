package dto

type CreatePlayerRequest struct {
	Name     string  `json:"name" validate:"required"`
	Email    string  `json:"email" validate:"required"`
	Handicap float64 `json:"handicap" validate:"required"`
}

type PlayerResponse struct {
	ID       int64   `json:"id"`
	Name     string  `json:"name"`
	Email    string  `json:"email"`
	Handicap float64 `json:"handicap"`
}

type GetPlayerResponse struct {
	ID       int64   `json:"id"`
	Name     string  `json:"name"`
	Email    string  `json:"email"`
	Handicap float64 `json:"handicap"`
}

type UpdatePlayerRequest struct {
	Name     *string  `json:"name"`
	Email    *string  `json:"email"`
	Handicap *float64 `json:"handicap"`
}

// NEW objects for Wolf game
// PlayerRoleResponse represents one player's participation in a game,
// unified across every game type. team is populated for two-sided
// formats (points_play/match_play/team_event matches); seat is
// populated for wolf's fixed rotation. Exactly one of the two is set,
// depending on game_type.
type PlayerRoleResponse struct {
	PlayerID int64   `json:"player_id"`
	Team     *string `json:"team,omitempty"`
	Seat     *int    `json:"seat,omitempty"`
	Name     string  `json:"name"`
	Email    string  `json:"email"`
	Handicap float64 `json:"handicap"`
}
