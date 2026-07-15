package game

import (
	"golf-game-kaffip/internal/domain/player"
	"testing"
)

func TestGrossBonus(t *testing.T) {
	tests := []struct {
		name  string
		gross int
		par   int
		want  int
	}{
		{"par is no bonus", 4, 4, 0},
		{"bogey is no bonus", 5, 4, 0},
		{"birdie", 3, 4, 1},
		{"eagle", 2, 4, 2},
		{"albatross caps at eagle's 2 points", 1, 4, 2},
		{"birdie on par 3", 2, 3, 1},
		{"eagle on par 3 (hole-in-one)", 1, 3, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := grossBonus(tt.gross, tt.par)
			if got != tt.want {
				t.Errorf("grossBonus(%d, %d) = %d, want %d", tt.gross, tt.par, got, tt.want)
			}
		})
	}
}

func TestLowestScoreWinner(t *testing.T) {
	teamOf := map[int64]string{1: "A", 2: "A", 3: "B", 4: "B"}

	tests := []struct {
		name   string
		scores map[int64]int
		want   string
	}{
		{
			name:   "team A player has exclusive lowest",
			scores: map[int64]int{1: 3, 2: 8, 3: 4, 4: 6},
			want:   "A",
		},
		{
			name:   "team B player has exclusive lowest",
			scores: map[int64]int{1: 5, 2: 6, 3: 3, 4: 7},
			want:   "B",
		},
		{
			name:   "cross-team tie for lowest awards no point",
			scores: map[int64]int{1: 4, 2: 8, 3: 4, 4: 6},
			want:   "",
		},
		{
			name:   "same-team tie for lowest still awards that team the point once",
			scores: map[int64]int{1: 4, 2: 4, 3: 5, 4: 6},
			want:   "A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := lowestScoreWinner(tt.scores, teamOf)
			if got != tt.want {
				t.Errorf("lowestScoreWinner() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTeamAccumulativeWinner(t *testing.T) {
	teamOf := map[int64]string{1: "A", 2: "A", 3: "B", 4: "B"}

	tests := []struct {
		name   string
		scores map[int64]int
		want   string
	}{
		{
			name:   "team A lower total wins",
			scores: map[int64]int{1: 3, 2: 3, 3: 4, 4: 4}, // A=6, B=8
			want:   "A",
		},
		{
			name:   "team B lower total wins",
			scores: map[int64]int{1: 3, 2: 8, 3: 4, 4: 6}, // A=11, B=10
			want:   "B",
		},
		{
			name:   "equal totals award no point",
			scores: map[int64]int{1: 4, 2: 5, 3: 5, 4: 4}, // A=9, B=9
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := teamAccumulativeWinner(tt.scores, teamOf)
			if got != tt.want {
				t.Errorf("teamAccumulativeWinner() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTallyHolePoints(t *testing.T) {
	tests := []struct {
		name            string
		lowScoreWinner  string
		teamTotalWinner string
		bonusPoints     map[string]int
		wantA, wantB    int
	}{
		{
			name:            "your worked example: A wins low+bonus, B wins total",
			lowScoreWinner:  "A",
			teamTotalWinner: "B",
			bonusPoints:     map[string]int{"A": 1, "B": 0},
			wantA:           2,
			wantB:           1,
		},
		{
			name:            "both categories tied, no bonuses",
			lowScoreWinner:  "",
			teamTotalWinner: "",
			bonusPoints:     map[string]int{"A": 0, "B": 0},
			wantA:           0,
			wantB:           0,
		},
		{
			name:            "team B sweeps everything",
			lowScoreWinner:  "B",
			teamTotalWinner: "B",
			bonusPoints:     map[string]int{"A": 0, "B": 2},
			wantA:           0,
			wantB:           4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotA, gotB := tallyHolePoints(tt.lowScoreWinner, tt.teamTotalWinner, tt.bonusPoints)
			if gotA != tt.wantA || gotB != tt.wantB {
				t.Errorf("tallyHolePoints() = (%d, %d), want (%d, %d)", gotA, gotB, tt.wantA, tt.wantB)
			}
		})
	}
}

func TestCalculateHoleResult(t *testing.T) {
	players := map[int64]*player.Player{
		1: {ID: 1, Handicap: 0},
		2: {ID: 2, Handicap: 0},
		3: {ID: 3, Handicap: 0},
		4: {ID: 4, Handicap: 0},
	}

	hole := HoleInfo{Number: 1, Par: 4, StrokeIndex: 9}
	const totalHoles = 18

	t.Run("your worked example, variant gross", func(t *testing.T) {
		inputs := []PlayerScoreInput{
			{PlayerID: 1, Gross: 3, TeamID: "A"},
			{PlayerID: 2, Gross: 8, TeamID: "A"},
			{PlayerID: 3, Gross: 4, TeamID: "B"},
			{PlayerID: 4, Gross: 6, TeamID: "B"},
		}

		result, err := calculateHoleResult(hole, inputs, VariantGross, players, totalHoles)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.PointsA != 2 || result.PointsB != 1 {
			t.Errorf("PointsA/B = %d/%d, want 2/1", result.PointsA, result.PointsB)
		}
		if result.LowScoreWinnerTeam != "A" {
			t.Errorf("LowScoreWinnerTeam = %q, want A", result.LowScoreWinnerTeam)
		}
		if result.TeamTotalWinnerTeam != "B" {
			t.Errorf("TeamTotalWinnerTeam = %q, want B", result.TeamTotalWinnerTeam)
		}
		if len(result.GrossBonuses) != 1 || result.GrossBonuses[0].PlayerID != 1 || result.GrossBonuses[0].Bonus != 1 {
			t.Errorf("GrossBonuses = %+v, want single birdie for player 1", result.GrossBonuses)
		}
	})

	t.Run("your correction example: cross-team tie nullifies low score category", func(t *testing.T) {
		inputs := []PlayerScoreInput{
			{PlayerID: 1, Gross: 4, TeamID: "A"},
			{PlayerID: 2, Gross: 8, TeamID: "A"},
			{PlayerID: 3, Gross: 4, TeamID: "B"},
			{PlayerID: 4, Gross: 6, TeamID: "B"},
		}

		result, err := calculateHoleResult(hole, inputs, VariantGross, players, totalHoles)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.PointsA != 0 || result.PointsB != 1 {
			t.Errorf("PointsA/B = %d/%d, want 0/1", result.PointsA, result.PointsB)
		}
		if result.LowScoreWinnerTeam != "" {
			t.Errorf("LowScoreWinnerTeam = %q, want empty (cross-team tie)", result.LowScoreWinnerTeam)
		}
		if len(result.GrossBonuses) != 0 {
			t.Errorf("GrossBonuses = %+v, want none (gross 4 = par, not birdie)", result.GrossBonuses)
		}
	})

	t.Run("variant net: handicap strokes shift the scoring basis but never the birdie check", func(t *testing.T) {
		netPlayers := map[int64]*player.Player{
			1: {ID: 1, Handicap: 18}, // gets exactly 1 stroke on every hole over 18 holes
			2: {ID: 2, Handicap: 0},
			3: {ID: 3, Handicap: 0},
			4: {ID: 4, Handicap: 0},
		}

		// Player 1 scores gross 4 (par) but nets to 3 via handicap stroke.
		inputs := []PlayerScoreInput{
			{PlayerID: 1, Gross: 4, TeamID: "A"},
			{PlayerID: 2, Gross: 4, TeamID: "A"},
			{PlayerID: 3, Gross: 4, TeamID: "B"},
			{PlayerID: 4, Gross: 4, TeamID: "B"},
		}

		result, err := calculateHoleResult(hole, inputs, VariantNet, netPlayers, totalHoles)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.LowScoreWinnerTeam != "A" {
			t.Errorf("LowScoreWinnerTeam = %q, want A (player 1 nets to 3 via handicap stroke)", result.LowScoreWinnerTeam)
		}

		if len(result.GrossBonuses) != 0 {
			t.Errorf("GrossBonuses = %+v, want none — net score win must not trigger gross bonus", result.GrossBonuses)
		}

		var p1Result *PlayerHoleResult
		for i := range result.Scores {
			if result.Scores[i].PlayerID == 1 {
				p1Result = &result.Scores[i]
			}
		}
		if p1Result == nil {
			t.Fatal("expected a PlayerHoleResult for player 1")
		}
		if p1Result.Strokes != 1 || p1Result.Net != 3 {
			t.Errorf("player 1 Strokes/Net = %d/%d, want 1/3", p1Result.Strokes, p1Result.Net)
		}
	})

	t.Run("rejects wrong number of scores", func(t *testing.T) {
		inputs := []PlayerScoreInput{
			{PlayerID: 1, Gross: 4, TeamID: "A"},
		}
		_, err := calculateHoleResult(hole, inputs, VariantGross, players, totalHoles)
		if err == nil {
			t.Error("expected error for wrong number of scores, got nil")
		}
	})

	t.Run("rejects unknown player id", func(t *testing.T) {
		inputs := []PlayerScoreInput{
			{PlayerID: 999, Gross: 4, TeamID: "A"},
			{PlayerID: 2, Gross: 4, TeamID: "A"},
			{PlayerID: 3, Gross: 4, TeamID: "B"},
			{PlayerID: 4, Gross: 4, TeamID: "B"},
		}
		_, err := calculateHoleResult(hole, inputs, VariantGross, players, totalHoles)
		if err == nil {
			t.Error("expected error for unknown player id, got nil")
		}
	})
}
