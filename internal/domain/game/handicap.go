package game

import (
	"golf-game-kaffip/internal/domain/player"
	"math"
)

// strokesReceived computes how many handicap strokes a player receives on
// a single hole, using standard stroke allocation: every player gets
// floor(handicap / totalHoles) strokes on every hole, plus one additional
// stroke on the hardest holes (lowest StrokeIndex) for the remainder.
//
// Example: handicap 20, totalHoles 18 → 1 extra stroke on every hole
// (20/18 = 1 rem 2), plus 1 more stroke on the 2 hardest holes
// (StrokeIndex 1 and 2) → those two holes get 2 strokes, the rest get 1.
func strokesReceived(handicap float64, strokeIndex int, totalHoles int) int {
	if totalHoles <= 0 {
		return 0
	}

	h := int(math.Round(handicap))
	if h < 0 {
		h = 0
	}

	base := h / totalHoles
	remainder := h % totalHoles

	strokes := base
	if strokeIndex <= remainder {
		strokes++
	}

	return strokes
}

// computeStartingLead implements the rudimentary Variant Gross handicap
// allowance: the team with the higher combined handicap starts with a
// lead equal to the rounded difference in combined handicaps.
// Positive → favors TeamA. Negative → favors TeamB. Zero → even.
func computeStartingLead(teamA, teamB []*player.Player) int {
	var combinedA, combinedB float64
	for _, p := range teamA {
		combinedA += p.Handicap
	}
	for _, p := range teamB {
		combinedB += p.Handicap
	}

	diff := combinedA - combinedB
	return int(math.Round(diff))
}
