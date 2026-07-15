package player

import "time"

type Player struct {
	ID        int64      `json:"id"`
	Name      string     `json:"name"`
	Handicap  float64    `json:"handicap"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

type UpdatePlayerParams struct {
	Name     *string
	Handicap *float64
}
