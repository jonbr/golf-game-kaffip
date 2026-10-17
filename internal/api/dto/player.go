package dto

type CreatePlayerRequest struct {
	Name     string  `json:"name" validate:"required"`
	Email    string  `json:"email" validate:"required"`
	Handicap float64 `json:"handicap"`
}

type PlayerResponse struct {
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

// PlayerRoleResponse is one player's participation in a game. Seat is set
// only for Wolf's rotation-ordered players list; two-sided formats use
// the team_a / team_b arrays instead.
type PlayerRoleResponse struct {
	PlayerID int64   `json:"player_id"`
	Seat     *int    `json:"seat,omitempty"`
	Name     string  `json:"name"`
	Email    string  `json:"email"`
	Handicap float64 `json:"handicap"`
}
