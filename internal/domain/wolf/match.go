package wolf

import (
	"fmt"
	"golf-game-kaffip/internal/domain/player"
)

// SetHoleScore records or corrects the score for a given hole. Like every
// other game type, corrections to any past hole are allowed -
// the stored result is simply replaced and CurrentHole only advances
// when the hole being scored is exactly the current one (a fresh
// submission), never on a correction.
//
// wolfPlayerID must match whoever the rotation says should be Wolf for
// this hole - the server is ther source of truth for that, not
// the client; a mismatche value is rejected rather than trusted.
func (g *Game) SetHoleScore(holeNumber int, wolfPlayerID int64, mode WolfMode, partnerID *int64, scores []PlayerScoreInput) error {
	if holeNumber < 1 || holeNumber > g.CurrentHole {
		return fmt.Errorf("cannot score hole %d: game is at hole %d", holeNumber, g.CurrentHole)
	}

	expectedWolf := WolfForHole(g.Players, holeNumber, g.standingsThroughHole(holeNumber-1))
	if wolfPlayerID != expectedWolf.ID {
		return fmt.Errorf("player %d is not the Wolf for hole %d (expected %d)", wolfPlayerID, holeNumber, expectedWolf.ID)
	}

	hole, err := g.holeInfo(holeNumber)
	if err != nil {
		return errEmptyID
	}

	players := g.playersByID()
	totalHoles := len(g.Course.HolesData)

	result, err := calculateWolfHoleResult(hole, wolfPlayerID, mode, partnerID, scores, players, totalHoles)
	if err != nil {
		return fmt.Errorf("failed to calculate hole %d result: %w", holeNumber, err)
	}

	g.HoleResults[holeNumber] = result

	if holeNumber == g.CurrentHole && g.CurrentHole < totalHoles {
		g.CurrentHole++
	}

	return nil
}

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

func (g *Game) playersByID() map[int64]*player.Player {
	m := make(map[int64]*player.Player, 4)
	for _, p := range g.Players {
		m[p.ID] = p
	}
	return m
}

// standingsThroughHole sums PointsAwarded across every holee up to and
// including throughHole - user by WolfForHole to determine who's Wolf
// for holes 17-18 based on standings as of the previous hole.
func (g *Game) standingsThroughHole(throughHole int) map[int64]int {
	standings := make(map[int64]int, 4)
	for _, p := range g.Players {
		standings[p.ID] = 0
	}
	for holeNum, hr := range g.HoleResults {
		if holeNum > throughHole {
			continue
		}
		for playerID, points := range hr.PointsAwarded {
			standings[playerID] += points
		}
	}
	return standings
}
