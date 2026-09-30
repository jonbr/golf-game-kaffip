package cup

import "time"

type Side string

const (
	SideA Side = "A"
	SideB Side = "B"
)

// Cup is a pure aggregator: several independent matches, grouped, with a
// roster fixing every participating player to one side. Cup holds no
// scoring logic of its own - see ComputeScore.
type Cup struct {
	ID         string
	Name       string
	Players    map[int64]Side // playerID -> side, fixed at creation
	MatchIDs   []string       // game IDs, in order
	CreatedAt  time.Time
	UpdatedAt  time.Time
	FinishedAt *time.Time
}

func NewCup(id, name string, players map[int64]Side, matchIDs []string) (*Cup, error) {
	if id == "" {
		return nil, errEmptyID
	}
	if len(players) == 0 {
		return nil, errNoPlayers
	}
	if len(matchIDs) == 0 {
		return nil, errNoMatches
	}

	hasA, hasB := false, false
	for _, s := range players {
		if s == SideA {
			hasA = true
		}
		if s == SideB {
			hasB = true
		}
	}
	if !hasA || !hasB {
		return nil, errIncompleteSides
	}

	return &Cup{
		ID:        id,
		Name:      name,
		Players:   players,
		MatchIDs:  matchIDs,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}
