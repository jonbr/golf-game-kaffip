package player

import "time"

type Player struct {
	ID        int64
	Name      string
	Email     string
	Handicap  float64
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type UpdatePlayerParams struct {
	Name     *string
	Email    *string
	Handicap *float64
}
