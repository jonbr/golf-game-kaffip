package application

import (
	"golf-game-kaffip/internal/domain/cup"
	domainGame "golf-game-kaffip/internal/domain/game"
	domainPlayer "golf-game-kaffip/internal/domain/player"
	domainWolf "golf-game-kaffip/internal/domain/wolf"
)

// resolveGameContribution determines a two-sided match's (point_play or
// match_play) contribution to the cup, using it's own Outcome(). Returns
// ok=false if the match isn't decided yet - undecided matches contribute
// nothing until finished.
func resolveGameContribution(g *domainGame.Game, roster map[int64]cup.Side) (cup.MatchContribution, bool) {
	outcome := g.Outcome()
	if !outcome.Decided {
		return cup.MatchContribution{}, false
	}

	c := cup.MatchContribution{GameID: g.ID}

	switch outcome.WinnerTeam {
	case "A":
		awardToRoster(&c, g.TeamA, roster, 1)
	case "B":
		awardToRoster(&c, g.TeamB, roster, 1)
	default: // halved
		awardToRoster(&c, g.TeamA, roster, 0.5)
		awardToRoster(&c, g.TeamB, roster, 0.5)
	}

	return c, true
}

// resolveWolfContribution determines a Wolf match's contribution: the
func resolveWolfContribution(g *domainWolf.Game, standings map[int64]int, roster map[int64]cup.Side) (cup.MatchContribution, bool) {
	if g.FinishedAt == nil {
		return cup.MatchContribution{}, false
	}

	topScore := -1
	for _, pts := range standings {
		if pts > topScore {
			topScore = pts
		}
	}

	var topPlayers []int64
	for _, p := range g.Players {
		if standings[p.ID] == topScore {
			topPlayers = append(topPlayers, p.ID)
		}
	}

	c := cup.MatchContribution{GameID: g.ID}

	if len(topPlayers) == 1 {
		awardToRosterIDs(&c, topPlayers, roster, 1)
		return c, true
	}

	// Tie for first: compare each side's combined total among ALL
	// participating plalyers (not just the tied ones).
	totalBySide := map[cup.Side]int{}
	for _, p := range g.Players {
		totalBySide[roster[p.ID]] += standings[p.ID]
	}

	switch {
	case totalBySide[cup.SideA] > totalBySide[cup.SideB]:
		c.PointsA = 1
	case totalBySide[cup.SideB] > totalBySide[cup.SideA]:
		c.PointsB = 1
	default:
		c.PointsA = 0.5
		c.PointsB = 0.5
	}

	return c, true
}

func awardToRoster(c *cup.MatchContribution, players []*domainPlayer.Player, roster map[int64]cup.Side, points float64) {
	ids := make([]int64, len(players))
	for i, p := range players {
		ids[i] = p.ID
	}
	awardToRosterIDs(c, ids, roster, points)
}

func awardToRosterIDs(c *cup.MatchContribution, playerIDs []int64, roster map[int64]cup.Side, points float64) {
	sides := map[cup.Side]bool{}
	for _, id := range playerIDs {
		sides[roster[id]] = true
	}
	for side := range sides {
		switch side {
		case cup.SideA:
			c.PointsA = points
		case cup.SideB:
			c.PointsB = points
		}
	}
}
