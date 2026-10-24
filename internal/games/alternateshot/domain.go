package alternateshot

import (
	"errors"
	"fmt"
	"golf-game-kaffip/internal/domain/course"
	domainGame "golf-game-kaffip/internal/domain/game"
	"golf-game-kaffip/internal/domain/player"
	"golf-game-kaffip/internal/domain/scoring"
	"time"
)

const GameType = domainGame.GameType("alternateshot")

type Game struct {
	ID          string
	Course      *course.Course
	Variant     domainGame.Variant
	TeamA       []*player.Player
	TeamB       []*player.Player
	CurrentHole int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	FinishedAt  *time.Time
	HoleResults map[int]*HoleResult
}

// HoleResult is one holes team-level result. Only one ball is played per
// side, so there is one score per team.
type HoleResult struct {
	Hole         domainGame.HoleInfo
	TeamAGross   int
	TeamBGross   int
	TeamAStrokes int
	TeamBStrokes int
	TeamANet     int
	TeamBNet     int
}

// WinerTeam return "A", "B" or "" if the hole is halved.
func (r *HoleResult) WinnerTeam() string {
	switch {
	case r.TeamANet < r.TeamBNet:
		return "A"
	case r.TeamBNet < r.TeamANet:
		return "B"
	default:
		return ""
	}
}

func (r *HoleResult) lead() int {
	switch r.WinnerTeam() {
	case "A":
		return 1
	case "B":
		return -1
	default:
		return 0
	}
}

func NewGame(id string, c *course.Course, teamA, teamB []*player.Player, variant domainGame.Variant) (*Game, error) {
	if id == "" {
		return nil, fmt.Errorf("game id cannot be empty")
	}
	if variant != domainGame.VariantGross && variant != domainGame.VariantNet {
		return nil, fmt.Errorf("invalid variant: %q", variant)
	}
	if len(teamA) != 2 || len(teamB) != 2 {
		return nil, fmt.Errorf("alternate shot requires exactly two players per side")
	}
	return &Game{
		ID:          id,
		Course:      c,
		Variant:     variant,
		TeamA:       teamA,
		TeamB:       teamB,
		CurrentHole: 1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		HoleResults: make(map[int]*HoleResult),
	}, nil
}

func (g *Game) TotalHoles() int { return len(g.Course.HolesData) }

func (g *Game) HoleInfo(holeNumber int) (domainGame.HoleInfo, error) {
	for _, h := range g.Course.HolesData {
		if h.Number == holeNumber {
			return domainGame.HoleInfo{
				Number:      h.Number,
				Par:         h.Par,
				StrokeIndex: h.HandicapIndex,
			}, nil
		}
	}
	return domainGame.HoleInfo{}, fmt.Errorf("hole %d not found on course %q", holeNumber, g.Course.ID)
}

// CalculateHoleResult turns two team gross scores into a HoleResult. Net
// games use 50% of the combined team handicap, spread by stroke index.
func (g *Game) CalculateHoleResult(hole domainGame.HoleInfo, teamAGross, teamBGross int) (*HoleResult, error) {
	if teamAGross < 1 || teamBGross < 1 {
		return nil, errors.New("gross scores must be at least 1")
	}
	r := &HoleResult{
		Hole:       hole,
		TeamAGross: teamAGross,
		TeamBGross: teamBGross,
		TeamANet:   teamAGross,
		TeamBNet:   teamBGross,
	}
	if g.Variant == domainGame.VariantNet {
		r.TeamAStrokes = scoring.StrokesReceived(teamHandicap(g.TeamA), hole.StrokeIndex, g.TotalHoles())
		r.TeamBStrokes = scoring.StrokesReceived(teamHandicap(g.TeamB), hole.StrokeIndex, g.TotalHoles())
		r.TeamANet = teamAGross - r.TeamAStrokes
		r.TeamBNet = teamBGross - r.TeamBStrokes
	}
	return r, nil
}

func (g *Game) SetHoleScore(holeNumber int, result *HoleResult) error {
	if holeNumber < 1 || holeNumber > g.CurrentHole {
		return fmt.Errorf("cannot score hole %d: game is at hole %d", holeNumber, g.CurrentHole)
	}
	g.HoleResults[holeNumber] = result
	if holeNumber == g.CurrentHole && g.CurrentHole < g.TotalHoles() {
		g.CurrentHole++
	}
	return nil
}

// Status is derived fresh from stored results, never stored.
func (g *Game) Status() domainGame.MatchPlayStatus {
	lead, lastPlayed := 0, 0
	for h := 1; h <= g.CurrentHole; h++ {
		r, ok := g.HoleResults[h]
		if !ok {
			continue
		}
		lead += r.lead()
		lastPlayed = h
	}
	return scoring.MatchStatusFromLead(lead, lastPlayed, g.TotalHoles())
}

func teamHandicap(team []*player.Player) float64 {
	sum := 0.0
	for _, p := range team {
		sum += p.Handicap
	}
	return sum * 0.5 // alternate shot uses 50% of combined team handicap
}

//func strokesOnHole(handicap, strokIndex, totalHoles int) int {}
