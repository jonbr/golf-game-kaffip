package cup

// MatchContribution is how much of a single match's 1 point each side
// earned - already resolved by whichever format produced it (game's
// Outcome() for two-sided matches, Wolf's highest-scorer rule for Wolf
// matches). Cup never computes this itself, only sums it.
type MatchContribution struct {
	GameID  string
	PointsA float64
	PointsB float64
}

type Score struct {
	TeamA float64
	TeamB float64
}

// ComputeScore sums pre-resolved per-match contributions. Always
// recompute fresh from current match state - never stored.
func ComputeScore(contributions []MatchContribution) Score {
	var s Score
	for _, c := range contributions {
		s.TeamA += c.PointsA
		s.TeamB += c.PointsB
	}
	return s
}
