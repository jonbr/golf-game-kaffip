package game

import (
	"golf-game-kaffip/internal/domain/player"
	"testing"
)

func TestStrokesReceived(t *testing.T) {
	tests := []struct {
		name        string
		handicap    float64
		strokeIndex int
		totalHoles  int
		want        int
	}{
		{"zero handicap, no strokes anywhere", 0, 1, 18, 0},
		{"handicap equal to total holes, one stroke everywhere", 18, 1, 18, 1},
		{"handicap equal to total holes, hardest hole", 18, 18, 18, 1},
		{"handicap under 18, hardest hole gets stroke", 5, 1, 18, 1},
		{"handicap under 18, easiest hole gets none", 5, 18, 18, 0},
		{"handicap under 18, boundary hole gets stroke", 5, 5, 18, 1},
		{"handicap under 18, just past boundary gets none", 5, 6, 18, 0},
		{"handicap over 18, wraps to second stroke on hardest holes", 24, 1, 18, 2},
		{"handicap over 18, wraps to second stroke on 6th hardest", 24, 6, 18, 2},
		{"handicap over 18, no second stroke past remainder", 24, 7, 18, 1},
		{"fractional handicap rounds before allocating", 14.4, 1, 18, 1},
		{"fractional handicap rounds up at .5", 14.5, 1, 18, 1},
		{"negative handicap clamped to zero", -3, 1, 18, 0},
		{"zero total holes returns zero", 10, 1, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := strokesReceived(tt.handicap, tt.strokeIndex, tt.totalHoles)
			if got != tt.want {
				t.Errorf("strokesReceived(%v, %d, %d) = %d, want %d",
					tt.handicap, tt.strokeIndex, tt.totalHoles, got, tt.want)
			}
		})
	}
}

func TestComputeStartingLead(t *testing.T) {
	mkPlayer := func(handicap float64) *player.Player {
		return &player.Player{Handicap: handicap}
	}

	tests := []struct {
		name  string
		teamA []*player.Player
		teamB []*player.Player
		want  int
	}{
		{
			name:  "equal combined handicaps, no lead",
			teamA: []*player.Player{mkPlayer(10), mkPlayer(10)},
			teamB: []*player.Player{mkPlayer(8), mkPlayer(12)},
			want:  0,
		},
		{
			name:  "team A higher combined handicap, positive lead favors A",
			teamA: []*player.Player{mkPlayer(14), mkPlayer(10)}, // 24
			teamB: []*player.Player{mkPlayer(8), mkPlayer(4)},   // 12
			want:  12,
		},
		{
			name:  "team B higher combined handicap, negative lead favors B",
			teamA: []*player.Player{mkPlayer(8), mkPlayer(4)},   // 12
			teamB: []*player.Player{mkPlayer(14), mkPlayer(10)}, // 24
			want:  -12,
		},
		{
			name:  "fractional difference rounds to nearest int",
			teamA: []*player.Player{mkPlayer(24.7), mkPlayer(14.7)}, // 39.4
			teamB: []*player.Player{mkPlayer(4.2), mkPlayer(34.2)},  // 38.4
			want:  1,                                                // diff 1.0 -> 1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeStartingLead(tt.teamA, tt.teamB)
			if got != tt.want {
				t.Errorf("computeStartingLead() = %d, want %d", got, tt.want)
			}
		})
	}
}
