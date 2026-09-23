package wolf

import (
	"fmt"
	"golf-game-kaffip/internal/domain/player"
)

// strokesReceived mirrors game.strokesReceived exactly (standare handicap
// allocation). Duplicated here rather than importd cross-package, sicne
// wolf is a peer domain pakcage to game, not a dependant of it.
func strokesReceived(handciap float64, strokeIndex int, totalHoles int) int {
	if totalHoles <= 0 {
		return 0
	}
	h := int(handciap + 0.5)
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

// caclulateWolfHoleResult scores one hole. Always net (full handicap
// allocation) - Wolf has no gross variant.
func calculateWolfHoleResult(
	hole HoleInfo,
	wolfPlayerID int64,
	mode WolfMode,
	partnerID *int64,
	inputs []PlayerScoreInput,
	players map[int64]*player.Player,
	totalHoles int,
) (*HoleResult, error) {
	if len(inputs) != 4 {
		return nil, fmt.Errorf("expected exactly 4 player scores, got %d", len(inputs))
	}
	if mode == WolfModePartnered && (partnerID == nil || *partnerID == wolfPlayerID) {
		return nil, fmt.Errorf("partner mode required a distinct partner id")
	}

	netByPlayer := make(map[int64]int, 4)
	results := make([]PlayerHoleResult, 0, 4)

	for _, in := range inputs {
		p, ok := players[in.PlayerID]
		if !ok {
			return nil, fmt.Errorf("unknown player id %d", in.PlayerID)
		}
		strokes := strokesReceived(p.Handicap, hole.StrokeIndex, totalHoles)
		net := in.Gross - strokes
		netByPlayer[in.PlayerID] = net
		results = append(results, PlayerHoleResult{
			PlayerID: in.PlayerID, Gross: in.Gross, Net: net, Strokes: strokes,
		})
	}

	wolfSideIDs := []int64{wolfPlayerID}
	if mode == WolfModePartnered {
		wolfSideIDs = append(wolfSideIDs, *partnerID)
	}

	var fieldSideIDs []int64
	for _, in := range inputs {
		isWolfSide := false
		for _, id := range wolfSideIDs {
			if in.PlayerID == id {
				isWolfSide = true
				break
			}
		}
		if !isWolfSide {
			fieldSideIDs = append(fieldSideIDs, in.PlayerID)
		}
	}

	wolfSideScore := bestNet(netByPlayer, wolfSideIDs)
	fieldSideScore := bestNet(netByPlayer, fieldSideIDs)

	points := make(map[int64]int, 4)
	for _, in := range inputs {
		points[in.PlayerID] = 0
	}

	var winningSide WolfSide
	switch {
	case wolfSideScore < fieldSideScore:
		winningSide = WolfSideWolf
		if mode == WolfModeLone {
			partnerID = nil
			points[wolfPlayerID] = 3
		} else {
			for _, id := range wolfSideIDs {
				points[id] = 1
			}
		}
	case fieldSideScore < wolfSideScore:
		winningSide = WolfSideField
		for _, id := range fieldSideIDs {
			points[id] = 1
		}
	default:
		// tied
		if mode == WolfModeLone {
			points[wolfPlayerID] = 1
		}
	}

	return &HoleResult{
		Hole:          hole,
		WolfPlayerID:  wolfPlayerID,
		Mode:          mode,
		PartnerID:     partnerID,
		Scores:        results,
		WinningSide:   winningSide,
		PointsAwarded: points,
	}, nil
}

func bestNet(neByPlayer map[int64]int, ids []int64) int {
	best := neByPlayer[ids[0]]
	for _, id := range ids[1:] {
		if neByPlayer[id] < best {
			best = neByPlayer[id]
		}
	}
	return best
}
