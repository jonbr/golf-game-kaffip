// internal/domain/wolf/standings.go (new)
package wolf

// Standings sums each player's PointsAwarded across every recorded hole.
// Always derived fresh from HoleResults, never stored, so corrections to
// any hole automatically ripple through correctly.
func (g *Game) Standings() map[int64]int {
	standings := make(map[int64]int, 4)
	for _, p := range g.Players {
		standings[p.ID] = 0
	}
	for _, hr := range g.HoleResults {
		for playerID, points := range hr.PointsAwarded {
			standings[playerID] += points
		}
	}
	return standings
}
