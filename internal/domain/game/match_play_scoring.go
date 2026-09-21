package game

import "fmt"

// MatchPlayStatus describes the current or final state of a match play
// game in classic terms (e.g. "3 up", "5&4", "Halved").
type MatchPlayStatus struct {
	WinnerTeam string // "A", "B", or "" if undecided or halved
	Display    string
	Closed     bool // true if the match is mathematically decided early
}

// MatchPlayStatus computes the current match play status by replaying
// stored hole results. It's always derived fresh, never stored, so it
// stays accurate even as players continue recording holes after the
// match is decided (allowed, since the app doubles as a bookkeeping tool).
func (g *Game) MatchPlayStatus() MatchPlayStatus {
	lastPlayed := 0
	lead := 0
	for h := 1; h <= g.CurrentHole; h++ {
		r, ok := g.HoleResults[h]
		if !ok {
			continue
		}
		lead += r.PointsA - r.PointsB
		lastPlayed = h
	}

	totalHoles := len(g.Course.HolesData)
	remaining := totalHoles - lastPlayed
	absLead := lead
	if absLead < 0 {
		absLead = -absLead
	}

	winner := ""
	if lead > 0 {
		winner = "A"
	} else if lead < 0 {
		winner = "B"
	}

	switch {
	case absLead > remaining && remaining > 0:
		return MatchPlayStatus{WinnerTeam: winner, Display: fmt.Sprintf("%d&%d", absLead, remaining), Closed: true}
	case remaining == 0 && lead == 0:
		return MatchPlayStatus{WinnerTeam: "", Display: "Halved", Closed: true}
	case remaining == 0:
		return MatchPlayStatus{WinnerTeam: winner, Display: fmt.Sprintf("%d up", absLead), Closed: true}
	case lead == 0:
		return MatchPlayStatus{WinnerTeam: "", Display: fmt.Sprintf("All Square thru %d", lastPlayed)}
	default:
		return MatchPlayStatus{WinnerTeam: winner, Display: fmt.Sprintf("%d up thru %d", absLead, lastPlayed)}
	}
}
