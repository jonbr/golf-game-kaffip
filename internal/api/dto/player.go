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

// PlayerRoleResponse represents one player's participation in a game.
// Seat is populated only for Wolf's flat, rotation-ordered player list
// (its fixed 0-3 seat order). Match Play and Team Points don't need a
// discriminator field here at all — their responses use TeamA/TeamB
// arrays, so which side a player is on is the array they appear in.
type PlayerRoleResponse struct {
	PlayerID int64   `json:"player_id"`
	Seat     *int    `json:"seat,omitempty"`
	Name     string  `json:"name"`
	Email    string  `json:"email"`
	Handicap float64 `json:"handicap"`
}
