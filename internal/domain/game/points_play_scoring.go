package game

import (
	"fmt"
	"golf-game-kaffip/internal/domain/player"
	"math"
)

// PlayerScoreInput is the raw input for one player's performance on a hole.
type PlayerScoreInput struct {
	PlayerID int64
	Gross    int
	TeamID   string // "A" or "B"
}

// TODO: Rename function since games types are increasing!
// calculateHoleResult computes the full HoleResult for a single hole, given
// the hole's info, the four players' gross scores, the game's variant, and
// the course's total hole count (needed for handicap stroke allocation).
func calculateHoleResult(hole HoleInfo, inputs []PlayerScoreInput, variant Variant, players map[int64]*player.Player, totalHoles int) (*HoleResult, error) {
	if len(inputs) != 4 {
		return nil, fmt.Errorf("expected exactly 4 player scores, got %d", len(inputs))
	}

	results, scoringBasis, teamOf, err := buildPlayerResults(hole, inputs, variant, players, totalHoles)
	if err != nil {
		return nil, err
	}

	lowScoreWinner := lowestScoreWinner(scoringBasis, teamOf)
	teamTotalWinner := teamAccumulativeWinner(scoringBasis, teamOf)
	bonuses, bonusPointsByTeam := collectGrossBonuses(results, teamOf)

	pointsA, pointsB := tallyHolePoints(lowScoreWinner, teamTotalWinner, bonusPointsByTeam)

	return &HoleResult{
		Hole:                hole,
		Scores:              results,
		GrossBonuses:        bonuses,
		LowScoreWinnerTeam:  lowScoreWinner,
		TeamTotalWinnerTeam: teamTotalWinner,
		PointsA:             pointsA,
		PointsB:             pointsB,
	}, nil
}

// calculateMatchPlayHoleResult scores a single hole for 1v1 match play:
// lower net score wins the hole (1 point), equal net scores halve it
// (0 points either side). Handicap strokes always apply, and there are
// no birdie/eagle bonus categories in this format.
func calculateMatchPlayHoleResult(hole HoleInfo, inputs []PlayerScoreInput, variant Variant, players map[int64]*player.Player, totalHoles int) (*HoleResult, error) {
	if len(inputs) != 2 {
		return nil, fmt.Errorf("expected exactly 2 player scores, got %d", len(inputs))
	}

	results := make([]PlayerHoleResult, 0, 2)
	scoringBasisByTeam := make(map[string]int, 2)

	for _, in := range inputs {
		p, ok := players[in.PlayerID]
		if !ok {
			return nil, fmt.Errorf("unknown player id %d", in.PlayerID)
		}

		strokes := 0
		if variant == VariantNet {
			strokes = strokesReceived(p.Handicap, hole.StrokeIndex, totalHoles)
		}
		net := in.Gross - strokes

		scoringBasisByTeam[in.TeamID] = net

		results = append(results, PlayerHoleResult{
			PlayerID: in.PlayerID,
			Gross:    in.Gross,
			Net:      net,
			Strokes:  strokes,
		})
	}

	pointsA, pointsB := 0, 0
	winner := ""
	switch {
	case scoringBasisByTeam["A"] < scoringBasisByTeam["B"]:
		pointsA = 1
		winner = "A"
	case scoringBasisByTeam["B"] < scoringBasisByTeam["A"]:
		pointsB = 1
		winner = "B"
	}

	return &HoleResult{
		Hole:               hole,
		Scores:             results,
		LowScoreWinnerTeam: winner,
		PointsA:            pointsA,
		PointsB:            pointsB,
	}, nil
}

// buildPlayerResults computes each player's net score, handicap strokes,
// and birdie/eagle bonus, returning the per-player results plus lookup
// maps used by the category functions below.
func buildPlayerResults(hole HoleInfo, inputs []PlayerScoreInput, variant Variant, players map[int64]*player.Player, totalHoles int) (
	results []PlayerHoleResult,
	scoringBasis map[int64]int,
	teamOf map[int64]string,
	err error,
) {
	scoringBasis = make(map[int64]int)
	teamOf = make(map[int64]string)

	for _, in := range inputs {
		p, ok := players[in.PlayerID]
		if !ok {
			return nil, nil, nil, fmt.Errorf("unknown player id %d", in.PlayerID)
		}

		strokes := 0
		if variant == VariantNet {
			strokes = strokesReceived(p.Handicap, hole.StrokeIndex, totalHoles)
		}
		net := in.Gross - strokes

		scoringBasis[in.PlayerID] = net
		teamOf[in.PlayerID] = in.TeamID

		results = append(results, PlayerHoleResult{
			PlayerID:   in.PlayerID,
			Gross:      in.Gross,
			Net:        net,
			Strokes:    strokes,
			GrossBonus: grossBonus(in.Gross, hole.Par),
		})
	}

	return results, scoringBasis, teamOf, nil
}

// grossBonus returns 1 for a birdie, 2 for an eagle or better, 0 otherwise.
// Always based on gross vs. par, regardless of variant.
func grossBonus(gross, par int) int {
	switch {
	case gross == par-1:
		return 1
	case gross <= par-2:
		return 2
	default:
		return 0
	}
}

// lowestScoreWinner returns "A" or "B" if one team exclusively holds the
// lowest score, or "" if the minimum is shared across both teams (tie).
func lowestScoreWinner(scores map[int64]int, teamOf map[int64]string) string {
	min := math.MaxInt
	for _, s := range scores {
		if s < min {
			min = s
		}
	}

	teamsAtMin := map[string]bool{}
	for pid, s := range scores {
		if s == min {
			teamsAtMin[teamOf[pid]] = true
		}
	}

	if len(teamsAtMin) == 1 {
		for team := range teamsAtMin {
			return team
		}
	}
	return ""
}

// teamAccumulativeWinner returns "A" or "B" if one team's summed score is
// strictly lower, or "" if the totals are equal (tie).
func teamAccumulativeWinner(scores map[int64]int, teamOf map[int64]string) string {
	totals := map[string]int{"A": 0, "B": 0}
	for pid, s := range scores {
		totals[teamOf[pid]] += s
	}

	switch {
	case totals["A"] < totals["B"]:
		return "A"
	case totals["B"] < totals["A"]:
		return "B"
	default:
		return ""
	}
}

// collectGrossBonuses gathers birdie/eagle bonuses already computed on
// each PlayerHoleResult into a flat list plus a per-team point total.
func collectGrossBonuses(results []PlayerHoleResult, teamOf map[int64]string) ([]GrossBonus, map[string]int) {
	var bonuses []GrossBonus
	totals := map[string]int{"A": 0, "B": 0}

	for _, r := range results {
		if r.GrossBonus > 0 {
			team := teamOf[r.PlayerID]
			bonuses = append(bonuses, GrossBonus{
				PlayerID: r.PlayerID,
				TeamID:   team,
				Bonus:    r.GrossBonus,
			})
			totals[team] += r.GrossBonus
		}
	}

	return bonuses, totals
}

// tallyHolePoints combines category 1, 2, and 3 results into each team's
// total points earned for the hole.
func tallyHolePoints(lowScoreWinner, teamTotalWinner string, bonusPoints map[string]int) (pointsA, pointsB int) {
	switch lowScoreWinner {
	case "A":
		pointsA++
	case "B":
		pointsB++
	}

	switch teamTotalWinner {
	case "A":
		pointsA++
	case "B":
		pointsB++
	}

	pointsA += bonusPoints["A"]
	pointsB += bonusPoints["B"]

	return pointsA, pointsB
}
