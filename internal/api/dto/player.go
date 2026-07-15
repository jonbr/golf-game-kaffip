package dto

type CreatePlayerRequest struct {
	Name     string  `json:"name" validate:"required"`
	Handicap float64 `json:"handicap" validate:"required"`
}

type CreatePlayerResponse struct {
	ID       int64   `json:"id"`
	Name     string  `json:"name"`
	Handicap float64 `json:"handicap"`
}

type GetPlayerResponse struct {
	ID       int64   `json:"id"`
	Name     string  `json:"name"`
	Handicap float64 `json:"handicap"`
}

type UpdatePlayerRequest struct {
	Name     *string  `json:"name"`
	Handicap *float64 `json:"handicap"`
}
