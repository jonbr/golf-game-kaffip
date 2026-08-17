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
			// diff = 12, 12/1.5 = 8.0 -> truncates to 8
			want: 8,
		},
		{
			name:  "team B higher combined handicap, negative lead favors B",
			teamA: []*player.Player{mkPlayer(8), mkPlayer(4)},   // 12
			teamB: []*player.Player{mkPlayer(14), mkPlayer(10)}, // 24
			// diff = -12, -12/1.5 = -8.0 -> truncates to -8
			want: -8,
		},
		{
			name:  "fractional diff truncates toward zero, not rounds",
			teamA: []*player.Player{mkPlayer(24.7), mkPlayer(14.7)}, // 39.4
			teamB: []*player.Player{mkPlayer(4.2), mkPlayer(34.2)},  // 38.4
			// diff = 1.0, 1.0/1.5 = 0.666... -> truncates to 0
			want: 0,
		},
		{
			name:  "larger fractional diff truncates down, not to nearest",
			teamA: []*player.Player{mkPlayer(30), mkPlayer(20)}, // 50
			teamB: []*player.Player{mkPlayer(10), mkPlayer(10)}, // 20
			// diff = 30, 30/1.5 = 20.0 exactly -> 20
			want: 20,
		},
		{
			name:  "negative fractional result truncates toward zero, not down",
			teamA: []*player.Player{mkPlayer(10)},
			teamB: []*player.Player{mkPlayer(15)},
			// diff = -5, -5/1.5 = -3.333... -> truncates to -3, not -4
			want: -3,
		},
		{
			name:  "empty team A returns zero without panicking",
			teamA: []*player.Player{},
			teamB: []*player.Player{mkPlayer(10)},
			want:  0,
		},
		{
			name:  "empty team B returns zero without panicking",
			teamA: []*player.Player{mkPlayer(10)},
			teamB: []*player.Player{},
			want:  0,
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
