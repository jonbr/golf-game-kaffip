package cup

import "testing"

func TestComputeSocre(t *testing.T) {
	tests := []struct {
		name          string
		contributions []MatchContribution
		wantA, wantB  float64
	}{
		{
			name:          "single match, clean win for A",
			contributions: []MatchContribution{{GameID: "g1", PointsA: 1, PointsB: 0}},
			wantA:         1, wantB: 0,
		},
		{
			name: "two matches, halved and a win, sum correctly",
			contributions: []MatchContribution{
				{GameID: "g1", PointsA: 0.5, PointsB: 0.5},
				{GameID: "g2", PointsA: 1, PointsB: 0},
			},
			wantA: 1.5, wantB: 0.5,
		},
		{
			contributions: nil,
			wantA:         0, wantB: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeScore(tt.contributions)
			if got.TeamA != tt.wantA || got.TeamB != tt.wantB {
				t.Errorf("ComputeScore() = %+v, want {%v %v}", got, tt.wantA, tt.wantB)
			}
		})
	}
}
