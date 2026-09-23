package wolf

import "golf-game-kaffip/internal/domain/player"

// WolfForHole determines which player is Wolf on a given hole. Holes 1-16
// follow the fixed rotation (seat = (hole-1) % 4). Holes 17-18 go to
// whoever has the fewest total points as of the end of the previous hole;
// ties are broken by earliest original rotation seat.
func WolfForHole(players [4]*player.Player, holeNumber int, standingsThroughPreviousHole map[int64]int) *player.Player {
	if holeNumber <= 16 {
		seat := (holeNumber - 1) % 4
		return players[seat]
	}

	lowestIdx := 0
	lowestPoints := standingsThroughPreviousHole[players[0].ID]
	for i := 1; i < 4; i++ {
		pts := standingsThroughPreviousHole[players[i].ID]
		if pts < lowestPoints {
			lowestPoints = pts
			lowestIdx = i
		}
	}
	return players[lowestIdx]
}
