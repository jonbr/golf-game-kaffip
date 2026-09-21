package wolf

import (
	"golf-game-kaffip/internal/domain/course"
	"golf-game-kaffip/internal/domain/player"
	"time"
)

type WolfMode string

const (
	WolfModePartnered WolfMode = "partnered"
	WolfModeLone      WolfMode = "lone"
)

type WolfSide string

const (
	WolfSideWolf  WolfSide = "wolf"
	WolfSideField WolfSide = "field"
)

type Game struct {
	ID          string
	Course      *course.Course
	Players     [4]*player.Player
	CurrentHole int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	FinishedAt  *time.Time
	HoleResults map[int]*HoleResult
}

type HoleResult struct {
	Hole          HoleInfo
	WolfPlayerID  int64
	Mode          WolfMode
	PartnerID     *int64
	Scores        []PlayerHoleResult
	WinningSide   WolfSide
	PointsAwarded map[int64]int
}

type PlayerHoleResult struct {
	PlayerID int64
	Gross    int
	Net      int
	Strokes  int
}

type HoleInfo struct {
	Number      int
	Par         int
	StrokeIndex int
}

type PlayerScoreInput struct {
	PlayerID int64
	Gross    int
}

func NewGame(id string, c *course.Course, players [4]*player.Player) (*Game, error) {
	if id == "" {
		return nil, errEmptyID
	}
	for _, p := range players {
		if p == nil {
			return nil, errMissingPlayer
		}
	}

	return &Game{
		ID:          id,
		Course:      c,
		Players:     players,
		CurrentHole: 1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		HoleResults: make(map[int]*HoleResult),
	}, nil
}
