package game

import (
	"fmt"
	"golf-game-kaffip/internal/domain/course"
	"golf-game-kaffip/internal/domain/player"
	"time"
)

// Variant determines how a game's scoring categories are computed.
type GameType string
type Variant string

const (
	GameTypeTeamPlay  GameType = "team_play"
	GameTypeMatchPlay GameType = "match_play"
)

const (
	VariantGross Variant = "gross" // no per-hole handicap; one-time starting lead instead
	VariantNet   Variant = "net"   // full per-hole handicap strokes, except birdies/eagles
)

type Game struct {
	ID           string
	Course       *course.Course
	GameType     GameType
	Variant      Variant
	TeamA        []*player.Player
	TeamB        []*player.Player
	CurrentHole  int
	CreatedAt    time.Time
	UpdatedAt    time.Time
	FinishedAt   *time.Time
	StartingLead int // signed: positive favors TeamA, negative favors TeamB
	MatchScore   MatchScore
	HoleResults  map[int]*HoleResult
}

// HoleResult is the domain representation of a scored hole.
type HoleResult struct {
	Hole                HoleInfo
	Scores              []PlayerHoleResult
	GrossBonuses        []GrossBonus
	LowScoreWinnerTeam  string // "A", "B", or "" (tie/no winner)
	TeamTotalWinnerTeam string // "A", "B", or "" (tie/no winner)
	PointsA             int    // total points TeamA earned this hole (all categories)
	PointsB             int    // total points TeamB earned this hole (all categories)
}

// PlayerHoleResult is the domain representation of a player's performance on a hole.
type PlayerHoleResult struct {
	PlayerID   int64
	Gross      int
	Net        int
	Strokes    int // handicap strokes received on this hole (0 for Variant Gross)
	GrossBonus int // 0, 1 (birdie), or 2 (eagle)
}

// HoleInfo is already in your domain, but if not:
type HoleInfo struct {
	Number      int
	Par         int
	StrokeIndex int
}

// MatchScore is the display projection of the running signed lead:
// the trailing team always shows 0.
type MatchScore struct {
	TeamA int
	TeamB int
}

type GrossBonus struct {
	PlayerID int64
	TeamID   string // "A" or "B"
	Bonus    int
}

// GameSummary is a lightweight projection of a Game, used for list views
// (GetGames). It intentionally excludes hole-by-hole results and full
// player rosters, which are expensive to assemble for many games at once
// and are only needed on the single-game detail path (LoadGame).
type GameSummary struct {
	ID          string
	CourseID    string
	CourseName  string
	GameType    GameType
	Variant     Variant
	CurrentHole int
	TotalHoles  int
	MatchScore  MatchScore
	CreatedAt   time.Time
	UpdatedAt   time.Time
	FinishedAt  *time.Time
}

func NewGame(
	id string,
	c *course.Course,
	teamA []*player.Player,
	teamB []*player.Player,
	gameType GameType,
	variant Variant,
) (*Game, error) {
	if id == "" {
		return nil, fmt.Errorf("game id cannot be empty")
	}

	if err := validateGameTypeAndVariant(gameType, variant, teamA, teamB); err != nil {
		return nil, err
	}

	startingLead := 0
	if variant == VariantGross {
		startingLead = computeStartingLead(teamA, teamB)
	}

	return &Game{
		ID:           id,
		Course:       c,
		GameType:     gameType,
		Variant:      variant,
		TeamA:        teamA,
		TeamB:        teamB,
		CurrentHole:  1,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		StartingLead: startingLead,
		HoleResults:  make(map[int]*HoleResult),
		MatchScore:   matchScoreFromLead(startingLead),
	}, nil
}

// validateGameTypeAndVariant enforces which (GameType, Variant, team size)
// combinations are currently supported. Adding a new supported
// combination later (e.g. match play + gross) is a change here, not a
// structural change elsewhere.
func validateGameTypeAndVariant(gameType GameType, variant Variant, teamA, teamB []*player.Player) error {
	if variant != VariantGross && variant != VariantNet {
		return fmt.Errorf("invalid variant: %q", variant)
	}

	switch gameType {
	case GameTypeTeamPlay:
		if len(teamA) != 2 || len(teamB) != 2 {
			return fmt.Errorf("team play requires exactly two players per side")
		}
	case GameTypeMatchPlay:
		if len(teamA) != 1 || len(teamB) != 1 {
			return fmt.Errorf("match play requires exactly one player per side")
		}
	default:
		return fmt.Errorf("invalid game type: %q", gameType)
	}
	return nil
}

func matchScoreFromLead(lead int) MatchScore {
	if lead > 0 {
		return MatchScore{TeamA: lead, TeamB: 0}
	}
	if lead < 0 {
		return MatchScore{TeamA: 0, TeamB: -lead}
	}
	return MatchScore{TeamA: 0, TeamB: 0}
}
