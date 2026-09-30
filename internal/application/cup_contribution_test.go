package application

import (
	"golf-game-kaffip/internal/domain/cup"
	domainGame "golf-game-kaffip/internal/domain/game"
	domainPlayer "golf-game-kaffip/internal/domain/player"
	domainWolf "golf-game-kaffip/internal/domain/wolf"
	"testing"
	"time"
)

func mkPlayer(id int64) *domainPlayer.Player {
	return &domainPlayer.Player{ID: id}
}

func TestResolveGameContribution(t *testing.T) {
	roster := map[int64]cup.Side{
		1: cup.SideA, 2: cup.SideB,
	}

	tests := []struct {
		name        string
		finished    bool
		matchScoreA int
		matchScoreB int
		wantOK      bool
		wantA       float64
		wantB       float64
	}{
		{"undecided, not finished", false, 0, 0, false, 0, 0},
		{"team A ahead and finished, A wins the point", true, 15, 0, true, 1, 0},
		{"team B ahead and finished, B wins the point", true, 0, 1, true, 0, 1},
		{"tied and finished, halved", true, 0, 0, true, 0.5, 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &domainGame.Game{
				ID:    "game_1",
				TeamA: []*domainPlayer.Player{mkPlayer(1)},
				TeamB: []*domainPlayer.Player{mkPlayer(2)},
				MatchScore: domainGame.MatchScore{
					TeamA: tt.matchScoreA, TeamB: tt.matchScoreB,
				},
			}
			if tt.finished {
				now := timeNow()
				g.FinishedAt = &now
			}

			c, ok := resolveGameContribution(g, roster)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if !ok {
				return
			}
			if c.PointsA != tt.wantA || c.PointsB != tt.wantB {
				t.Errorf("contribution = %+v, want A=%v B=%v", c, tt.wantA, tt.wantB)
			}
		})
	}
}

func TestResolveWolfcontribution(t *testing.T) {
	// 4 players: 1,2 on cup side A; 3,4 on cup side B.
	roster := map[int64]cup.Side{
		1: cup.SideA, 2: cup.SideA, 3: cup.SideB, 4: cup.SideB,
	}
	players := [4]*domainPlayer.Player{mkPlayer(1), mkPlayer(2), mkPlayer(3), mkPlayer(4)}

	tests := []struct {
		name      string
		standings map[int64]int
		wantA     float64
		wantB     float64
	}{
		{
			name:      "clean single highest scorer on side A",
			standings: map[int64]int{1: 5, 2: 2, 3: 1, 4: 0},
			wantA:     1, wantB: 0,
		},
		{
			name:      "clean single highest scorer on side B",
			standings: map[int64]int{1: 1, 2: 0, 3: 6, 4: 2},
			wantA:     0, wantB: 1,
		},
		{
			name: "tie for first between the two sides, fall back to combined totals - A wins",
			// tied at 5: player 1 (side A) and player 3 (side B)
			// side A combined: 5+3=8, side B combined: 5+1=6 -> A wins
			standings: map[int64]int{1: 5, 2: 3, 3: 5, 4: 1},
			wantA:     1, wantB: 0,
		},
		{
			name: "tie for first, combined totals also tied - halved",
			// tied at 5: player 1 (A) and player 3 (B)
			// side A combined: 5+2=7, side B combined: 5+2=7 -> halved
			standings: map[int64]int{1: 5, 2: 2, 3: 5, 4: 2},
			wantA:     0.5, wantB: 0.5,
		},
		{
			name: "tie for first WITHIN the same side, that side still wins outright",
			// tied at 5: player 1 and player 2, both side A
			standings: map[int64]int{1: 5, 2: 5, 3: 3, 4: 1},
			wantA:     1, wantB: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := timeNow()
			g := &domainWolf.Game{ID: "wolf_1", Players: players, FinishedAt: &now}

			c, ok := resolveWolfContribution(g, tt.standings, roster)
			if !ok {
				t.Fatalf("expected ok=true for a finished wolf match")
			}
			if c.PointsA != tt.wantA || c.PointsB != tt.wantB {
				t.Errorf("contribution = %+v, want A=%v B=%v", c, tt.wantA, tt.wantB)
			}
		})
	}
}

func TestResolveWolfContribution_NotFinished(t *testing.T) {
	players := [4]*domainPlayer.Player{mkPlayer(1), mkPlayer(2), mkPlayer(3), mkPlayer(4)}
	g := &domainWolf.Game{ID: "wolf_1", Players: players} // FinishedAt is nil

	_, ok := resolveWolfContribution(g, map[int64]int{}, map[int64]cup.Side{})
	if ok {
		t.Errorf("expected ok=false for an unfished wolf match")
	}
}

func timeNow() time.Time {
	return time.Now()
}
