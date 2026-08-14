package game

import "fmt"

// MatchOutcome is the game-type-agnostic result of a single match within
// a TeamEvent: who won it, whether it's decided yet, and a display string.
type MatchOutcome struct {
	WinnerTeam string // "A", "B", or "" if halved/undecided
	Decided    bool
	Display    string
}

// Outcome computes this game's outcome for event-aggregation purposes.
// A match only counts toward an event's score once it's been finished
// (FinishedAt set, via the existing FinishGame endpoint) — whoever is
// ahead at that point wins the match's point; an even score splits it
// 0.5/0.5. Works identically for match_play and team_play, since both
// already maintain MatchScore via the same replay logic.
func (g *Game) Outcome() MatchOutcome {
	if g.FinishedAt == nil {
		return MatchOutcome{Decided: false, Display: g.inProgressDisplay()}
	}

	winner := ""
	switch {
	case g.MatchScore.TeamA > g.MatchScore.TeamB:
		winner = "A"
	case g.MatchScore.TeamB > g.MatchScore.TeamA:
		winner = "B"
	}

	display := "Halved"
	if winner != "" {
		lead := g.MatchScore.TeamA
		if winner == "B" {
			lead = g.MatchScore.TeamB
		}
		if g.GameType == GameTypeMatchPlay {
			display = g.MatchPlayStatus().Display
		} else {
			display = fmt.Sprintf("%d pts", lead)
		}
	}

	return MatchOutcome{WinnerTeam: winner, Decided: true, Display: display}
}

func (g *Game) inProgressDisplay() string {
	if g.GameType == GameTypeMatchPlay {
		return g.MatchPlayStatus().Display
	}
	return fmt.Sprintf("thru %d", g.CurrentHole-1)
}
