package game

import (
	"fmt"
	"golf-game-kaffip/internal/domain/player"
)

// SetHoleScore stores an already-computed HoleResult for the given hole,
// then recomputes the match score and advances CurrentHole if applicable.
// The caller (a format-specific service) is responsible for computing
// the result via its own scoring logic — this method only applies it.
//
// Any hole from 1 up to CurrentHole can be targeted — including holes
// already played — since corrections are a first-class operation, not
// an exception.
func (g *Game) SetHoleScore(holeNumber int, result *HoleResult) error {
	if holeNumber < 1 || holeNumber > g.CurrentHole {
		return fmt.Errorf("cannot score hole %d: game is at hole %d", holeNumber, g.CurrentHole)
	}

	totalHoles := len(g.Course.HolesData)

	g.HoleResults[holeNumber] = result
	g.recomputeMatchScore()

	if holeNumber == g.CurrentHole && g.CurrentHole < totalHoles {
		g.CurrentHole++
	}

	return nil
}

// HoleInfo looks up a hole's par and stroke index from the course data.
// Exported since format-specific scoring packages need it to compute a
// HoleResult before calling SetHoleScore.
func (g *Game) HoleInfo(holeNumber int) (HoleInfo, error) {
	for _, h := range g.Course.HolesData {
		if h.Number == holeNumber {
			return HoleInfo{
				Number:      h.Number,
				Par:         h.Par,
				StrokeIndex: h.HandicapIndex,
			}, nil
		}
	}
	return HoleInfo{}, fmt.Errorf("hole %d not found on course %q", holeNumber, g.Course.ID)
}

// PlayersByID builds a lookup map of every player in the game, from both
// teams, keyed by player ID. Exported for the same reason as HoleInfo.
func (g *Game) PlayersByID() map[int64]*player.Player {
	m := make(map[int64]*player.Player, len(g.TeamA)+len(g.TeamB))
	for _, p := range g.TeamA {
		m[p.ID] = p
	}
	for _, p := range g.TeamB {
		m[p.ID] = p
	}
	return m
}

// TotalHoles returns the course's hole count.
func (g *Game) TotalHoles() int {
	return len(g.Course.HolesData)
}

func (g *Game) recomputeMatchScore() {
	lead := g.StartingLead

	for h := 1; h <= g.CurrentHole; h++ {
		result, ok := g.HoleResults[h]
		if !ok {
			continue
		}
		lead += result.PointsA - result.PointsB
	}

	g.MatchScore = matchScoreFromLead(lead)
}

// MatchPlayStatus describes the current or final state of a match play
// game in classic terms (e.g. "3 up", "5&4", "Halved").
type MatchPlayStatus struct {
	WinnerTeam string // "A", "B", or "" if undecided or halved
	Display    string
	Closed     bool // true if the match is mathematically decided early
}

// MatchPlayStatus computes the current match play status by replaying
// stored hole results. It's always derived fresh, never stored, so it
// stays accurate even as players continue recording holes after the
// match is decided (allowed, since the app doubles as a bookkeeping tool).
func (g *Game) MatchPlayStatus() MatchPlayStatus {
	lastPlayed := 0
	lead := 0
	for h := 1; h <= g.CurrentHole; h++ {
		r, ok := g.HoleResults[h]
		if !ok {
			continue
		}
		lead += r.PointsA - r.PointsB
		lastPlayed = h
	}

	totalHoles := len(g.Course.HolesData)
	remaining := totalHoles - lastPlayed
	absLead := lead
	if absLead < 0 {
		absLead = -absLead
	}

	winner := ""
	if lead > 0 {
		winner = "A"
	} else if lead < 0 {
		winner = "B"
	}

	switch {
	case absLead > remaining && remaining > 0:
		return MatchPlayStatus{WinnerTeam: winner, Display: fmt.Sprintf("%d&%d", absLead, remaining), Closed: true}
	case remaining == 0 && lead == 0:
		return MatchPlayStatus{WinnerTeam: "", Display: "Halved", Closed: true}
	case remaining == 0:
		return MatchPlayStatus{WinnerTeam: winner, Display: fmt.Sprintf("%d up", absLead), Closed: true}
	case lead == 0:
		return MatchPlayStatus{WinnerTeam: "", Display: fmt.Sprintf("All Square thru %d", lastPlayed)}
	default:
		return MatchPlayStatus{WinnerTeam: winner, Display: fmt.Sprintf("%d up thru %d", absLead, lastPlayed)}
	}
}
