package matchplay

import (
	"fmt"

	"golf-game-kaffip/internal/domain/game"
	"golf-game-kaffip/internal/domain/player"
)

// CalculateHoleResult scores a single hole for 1v1 match play: lower net
// score wins the hole (1 point), equal net scores halve it (0 points
// either side). Handicap strokes apply per variant; no birdie/eagle
// bonus categories in this format.
func CalculateHoleResult(hole game.HoleInfo, inputs []game.PlayerScoreInput, variant game.Variant, players map[int64]*player.Player, totalHoles int) (*game.HoleResult, error) {
	if len(inputs) != 2 {
		return nil, fmt.Errorf("expected exactly 2 player scores, got %d", len(inputs))
	}

	results := make([]game.PlayerHoleResult, 0, 2)
	scoringBasisByTeam := make(map[string]int, 2)

	for _, in := range inputs {
		p, ok := players[in.PlayerID]
		if !ok {
			return nil, fmt.Errorf("unknown player id %d", in.PlayerID)
		}

		strokes := 0
		if variant == game.VariantNet {
			strokes = strokesReceived(p.Handicap, hole.StrokeIndex, totalHoles)
		}
		net := in.Gross - strokes

		scoringBasisByTeam[in.TeamID] = net

		results = append(results, game.PlayerHoleResult{
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

	return &game.HoleResult{
		Hole:               hole,
		Scores:             results,
		LowScoreWinnerTeam: winner,
		PointsA:            pointsA,
		PointsB:            pointsB,
	}, nil
}
