package game

import (
	"testing"

	"golf-game-kaffip/internal/domain/course"
	"golf-game-kaffip/internal/domain/player"
)

func TestNewGame_StartingLeadOnlyAppliesToTeamPlayGross(t *testing.T) {
	c := &course.Course{
		ID: "test-course",
		HolesData: []course.Hole{
			{Number: 1, Par: 4, HandicapIndex: 1},
		},
	}

	// Deliberately unequal handicaps, so a non-zero surplus WOULD be
	// computed if computeStartingLead were mistakenly invoked.
	teamAHighHandicap := []*player.Player{{ID: 1, Handicap: 30}}
	teamBLowHandicap := []*player.Player{{ID: 2, Handicap: 5}}

	teamAHighHandicap2v2 := []*player.Player{{ID: 1, Handicap: 30}, {ID: 2, Handicap: 20}}
	teamBLowHandicap2v2 := []*player.Player{{ID: 3, Handicap: 5}, {ID: 4, Handicap: 5}}

	tests := []struct {
		name        string
		gameType    GameType
		variant     Variant
		teamA       []*player.Player
		teamB       []*player.Player
		wantNonZero bool
	}{
		{
			name:        "team play + gross gets a starting lead",
			gameType:    GameTypeTeamPoints,
			variant:     VariantGross,
			teamA:       teamAHighHandicap2v2,
			teamB:       teamBLowHandicap2v2,
			wantNonZero: true,
		},
		{
			name:        "team play + net gets no starting lead",
			gameType:    GameTypeTeamPoints,
			variant:     VariantNet,
			teamA:       teamAHighHandicap2v2,
			teamB:       teamBLowHandicap2v2,
			wantNonZero: false,
		},
		{
			name:        "match play + gross gets no starting lead",
			gameType:    GameTypeMatchPlay,
			variant:     VariantGross,
			teamA:       teamAHighHandicap,
			teamB:       teamBLowHandicap,
			wantNonZero: false,
		},
		{
			name:        "match play + net gets no starting lead",
			gameType:    GameTypeMatchPlay,
			variant:     VariantNet,
			teamA:       teamAHighHandicap,
			teamB:       teamBLowHandicap,
			wantNonZero: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := NewGame("test-game", c, tt.teamA, tt.teamB, tt.gameType, tt.variant)
			if err != nil {
				t.Fatalf("NewGame failed: %v", err)
			}

			gotNonZero := g.StartingLead != 0
			if gotNonZero != tt.wantNonZero {
				t.Errorf("StartingLead = %d, wantNonZero=%v", g.StartingLead, tt.wantNonZero)
			}
		})
	}
}
