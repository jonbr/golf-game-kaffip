package game

import (
	"fmt"
	"golf-game-kaffip/internal/domain/player"
)

// SetHoleScore records (or corrects) the score for a given hole. It can
// target any hole from 1 up to CurrentHole — including holes already
// played — since corrections are a first-class operation, not an
// exception. After storing the hole's result, the entire match score is
// recomputed by replaying every played hole in order, so corrections to
// earlier holes always ripple forward correctly.
func (g *Game) SetHoleScore(holeNumber int, scores []PlayerScoreInput) error {
	if holeNumber < 1 || holeNumber > g.CurrentHole {
		return fmt.Errorf("cannot score hole %d: game is at hole %d", holeNumber, g.CurrentHole)
	}

	hole, err := g.holeInfo(holeNumber)
	if err != nil {
		return err
	}

	players := g.playersByID()

	result, err := calculateHoleResult(hole, scores, g.Variant, players, len(g.Course.HolesData))
	if err != nil {
		return fmt.Errorf("failed to calculate hole %d result: %w", holeNumber, err)
	}

	g.HoleResults[holeNumber] = result
	g.recomputeMatchScore()

	if holeNumber == g.CurrentHole && g.CurrentHole < len(g.Course.HolesData) {
		g.CurrentHole++
	}

	return nil
}

// holeInfo looks up a hole's par and stroke index from the course data.
func (g *Game) holeInfo(holeNumber int) (HoleInfo, error) {
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

// playersByID builds a lookup map of every player in the game, from both
// teams, keyed by player ID.
func (g *Game) playersByID() map[int64]*player.Player {
	m := make(map[int64]*player.Player, len(g.TeamA)+len(g.TeamB))
	for _, p := range g.TeamA {
		m[p.ID] = p
	}
	for _, p := range g.TeamB {
		m[p.ID] = p
	}
	return m
}

// recomputeMatchScore rebuilds MatchScore from scratch by replaying every
// recorded hole result in order, starting from the game's initial lead
// (StartingLead for VariantGross, 0 for VariantNet). This is what makes
// correcting any past hole safe: since the lead is always fully
// recalculated rather than incrementally patched, a correction to an
// earlier hole correctly ripples through every hole that came after it.
func (g *Game) recomputeMatchScore() {
	lead := g.StartingLead

	for h := 1; h <= g.CurrentHole; h++ {
		result, ok := g.HoleResults[h]
		if !ok {
			continue // hole not yet scored
		}
		lead += result.PointsA - result.PointsB
	}

	g.MatchScore = matchScoreFromLead(lead)
}
