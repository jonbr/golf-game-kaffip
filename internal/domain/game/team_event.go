package game

import (
	"fmt"
	"time"

	"golf-game-kaffip/internal/domain/course"
)

// TeamEvent groups several 1v1 match play games into one Ryder-Cup-style
// team event. Each match is an ordinary Game with GameType=GameTypeMatchPlay
// — TeamEvent is purely an aggregation layer, it doesn't alter how an
// individual match is scored. Its aggregate score is always derived fresh
// from the current state of its matches, never stored, so it can't drift.
type TeamEvent struct {
	ID         string
	Course     *course.Course
	Variant    Variant
	MatchIDs   []string
	Matches    []*Game // GameType=GameTypeMatchPlay; TeamA[0]/TeamB[0] map directly to event sides
	CreatedAt  time.Time
	UpdatedAt  time.Time
	FinishedAt *time.Time
}

type EventScore struct {
	TeamA float64
	TeamB float64
}

// Score computes the aggregate score Ryder Cup style: 1 point to a decided
// match's winner, 0.5/0.5 if halved. Undecided matches contribute nothing.
func (e *TeamEvent) Score() EventScore {
	var score EventScore
	for _, m := range e.Matches {
		status := m.MatchPlayStatus()
		if !status.Closed {
			continue
		}
		switch status.WinnerTeam {
		case "A":
			score.TeamA += 1
		case "B":
			score.TeamB += 1
		default:
			score.TeamA += 0.5
			score.TeamB += 0.5
		}
	}
	return score
}

func NewTeamEvent(id string, c *course.Course, variant Variant, matches []*Game) (*TeamEvent, error) {
	if id == "" {
		return nil, fmt.Errorf("event id cannot be empty")
	}
	if variant != VariantGross && variant != VariantNet {
		return nil, fmt.Errorf("invalid variant: %q", variant)
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("event requires at least one match")
	}
	for _, m := range matches {
		if m.GameType != GameTypeMatchPlay {
			return nil, fmt.Errorf("event matches must be match_play games")
		}
	}

	matchIDs := make([]string, len(matches))
	for i, m := range matches {
		matchIDs[i] = m.ID
	}

	return &TeamEvent{
		ID:        id,
		Course:    c,
		Variant:   variant,
		MatchIDs:  matchIDs,
		Matches:   matches,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}
