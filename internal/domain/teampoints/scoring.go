package teampoints

import (
	"fmt"
	"math"

	"golf-game-kaffip/internal/domain/game"
	"golf-game-kaffip/internal/domain/player"
)

func CalculateHoleResult(hole game.HoleInfo, inputs []game.PlayerScoreInput, variant game.Variant, players map[int64]*player.Player, totalHoles int) (*game.HoleResult, error) {
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

	return &game.HoleResult{
		Hole:                hole,
		Scores:              results,
		GrossBonuses:        bonuses,
		LowScoreWinnerTeam:  lowScoreWinner,
		TeamTotalWinnerTeam: teamTotalWinner,
		PointsA:             pointsA,
		PointsB:             pointsB,
	}, nil
}

func buildPlayerResults(hole game.HoleInfo, inputs []game.PlayerScoreInput, variant game.Variant, players map[int64]*player.Player, totalHoles int) (
	results []game.PlayerHoleResult,
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
		if variant == game.VariantNet {
			strokes = strokesReceived(p.Handicap, hole.StrokeIndex, totalHoles)
		}
		net := in.Gross - strokes

		scoringBasis[in.PlayerID] = net
		teamOf[in.PlayerID] = in.TeamID

		results = append(results, game.PlayerHoleResult{
			PlayerID:   in.PlayerID,
			Gross:      in.Gross,
			Net:        net,
			Strokes:    strokes,
			GrossBonus: grossBonus(in.Gross, hole.Par),
		})
	}

	return results, scoringBasis, teamOf, nil
}

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

func collectGrossBonuses(results []game.PlayerHoleResult, teamOf map[int64]string) ([]game.GrossBonus, map[string]int) {
	var bonuses []game.GrossBonus
	totals := map[string]int{"A": 0, "B": 0}

	for _, r := range results {
		if r.GrossBonus > 0 {
			team := teamOf[r.PlayerID]
			bonuses = append(bonuses, game.GrossBonus{
				PlayerID: r.PlayerID,
				TeamID:   team,
				Bonus:    r.GrossBonus,
			})
			totals[team] += r.GrossBonus
		}
	}

	return bonuses, totals
}

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
