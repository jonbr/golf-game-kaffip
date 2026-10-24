package scoring

import "fmt"

// MatchStatus describes the current or final state of a match in classic
// terms (e.g. "3 up", "5&4", "Halved").
type MatchStatus struct {
	WinnerTeam string // "A", "B", or "" if undecided or halved
	Display    string
	Closed     bool // true if the match is mathematically decided early
}

// MatchStatusFromLead derives the status from a signed lead (positive favors
// team A) the highest hole with a result, and the course's total holes.
func MatchStatusFromLead(lead, lastPlayed, totalHoles int) MatchStatus {
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
		return MatchStatus{WinnerTeam: winner, Display: fmt.Sprintf("%d&%d", absLead, remaining), Closed: true}
	case remaining == 0 && lead == 0:
		return MatchStatus{WinnerTeam: "", Display: "Halved", Closed: true}
	case remaining == 0:
		return MatchStatus{WinnerTeam: winner, Display: fmt.Sprintf("%d up", absLead), Closed: true}
	case lead == 0:
		return MatchStatus{WinnerTeam: "", Display: fmt.Sprintf("All Square thru %d", lastPlayed)}
	default:
		return MatchStatus{WinnerTeam: winner, Display: fmt.Sprintf("%d up thru %d", absLead, lastPlayed)}
	}
}
