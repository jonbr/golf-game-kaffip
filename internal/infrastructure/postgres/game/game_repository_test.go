//go:build integration

package game_test

import (
	"context"
	"fmt"
	"testing"

	domainCourse "golf-game-kaffip/internal/domain/course"
	domainGame "golf-game-kaffip/internal/domain/game"
	domainPlayer "golf-game-kaffip/internal/domain/player"
	gamerepo "golf-game-kaffip/internal/infrastructure/postgres/game"
	playerrepo "golf-game-kaffip/internal/infrastructure/postgres/player"
	"golf-game-kaffip/internal/testutil"
)

func TestGameRepository_CreateGameAndLoadGame(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	ctx := context.Background()

	playerRepo := playerrepo.NewRepository(pool)
	gameRepo := gamerepo.NewRepository(pool, playerRepo)

	// Seed 4 real players, since game_players has a FK to players.
	players := make([]*domainPlayer.Player, 4)
	handicaps := []float64{24.7, 14.7, 4.2, 34.2}
	for i, h := range handicaps {
		p := &domainPlayer.Player{
			Name:     "Test Player",
			Email:    fmt.Sprintf("player%d@example.com", i),
			Handicap: h}
		if err := playerRepo.Create(ctx, p); err != nil {
			t.Fatalf("failed to seed player %d: %v", i, err)
		}
		players[i] = p
	}

	teamA := players[0:2]
	teamB := players[2:4]

	course := &domainCourse.Course{
		ID:   "test-course-id",
		Name: "Test Course",
		HolesData: []domainCourse.Hole{
			{Number: 1, Par: 4, HandicapIndex: 9},
			{Number: 2, Par: 3, HandicapIndex: 11},
			{Number: 3, Par: 5, HandicapIndex: 1},
		},
	}

	g, err := domainGame.NewGame("test-game-id", course, teamA, teamB, domainGame.GameTypeTeamPlay, domainGame.VariantGross)
	if err != nil {
		t.Fatalf("NewGame failed: %v", err)
	}

	if err := gameRepo.CreateGame(ctx, g); err != nil {
		t.Fatalf("CreateGame failed: %v", err)
	}

	loaded, err := gameRepo.LoadGame(ctx, "test-game-id")
	if err != nil {
		t.Fatalf("LoadGame failed: %v", err)
	}

	if loaded.ID != g.ID {
		t.Errorf("loaded.ID = %q, want %q", loaded.ID, g.ID)
	}
	if loaded.Course.Name != "Test Course" {
		t.Errorf("loaded.Course.Name = %q, want %q", loaded.Course.Name, "Test Course")
	}
	if len(loaded.Course.HolesData) != 3 {
		t.Fatalf("loaded.Course.HolesData has %d holes, want 3", len(loaded.Course.HolesData))
	}
	if loaded.Course.HolesData[0].Par != 4 || loaded.Course.HolesData[0].HandicapIndex != 9 {
		t.Errorf("loaded hole 1 = %+v, want Par=4 HandicapIndex=9", loaded.Course.HolesData[0])
	}
	if loaded.Variant != domainGame.VariantGross {
		t.Errorf("loaded.Variant = %q, want %q", loaded.Variant, domainGame.VariantGross)
	}
	if loaded.StartingLead != g.StartingLead {
		t.Errorf("loaded.StartingLead = %d, want %d", loaded.StartingLead, g.StartingLead)
	}
	if len(loaded.TeamA) != 2 || len(loaded.TeamB) != 2 {
		t.Fatalf("loaded teams = %d/%d, want 2/2", len(loaded.TeamA), len(loaded.TeamB))
	}
	if loaded.TeamA[0].ID != teamA[0].ID {
		t.Errorf("loaded.TeamA[0].ID = %d, want %d", loaded.TeamA[0].ID, teamA[0].ID)
	}
}

func TestGameRepository_SaveHoleResultAndReload(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	ctx := context.Background()

	playerRepo := playerrepo.NewRepository(pool)
	gameRepo := gamerepo.NewRepository(pool, playerRepo)

	players := make([]*domainPlayer.Player, 4)
	handicaps := []float64{0, 0, 0, 0}
	for i, h := range handicaps {
		p := &domainPlayer.Player{
			Name:     "Test Player",
			Email:    fmt.Sprintf("player%d@example.com", i),
			Handicap: h}
		if err := playerRepo.Create(ctx, p); err != nil {
			t.Fatalf("failed to seed player %d: %v", i, err)
		}
		players[i] = p
	}

	teamA := players[0:2]
	teamB := players[2:4]

	course := &domainCourse.Course{
		ID:   "test-course-id",
		Name: "Test Course",
		HolesData: []domainCourse.Hole{
			{Number: 1, Par: 4, HandicapIndex: 9},
			{Number: 2, Par: 3, HandicapIndex: 11},
			{Number: 3, Par: 5, HandicapIndex: 1},
		},
	}

	g, err := domainGame.NewGame("test-game-hole-result", course, teamA, teamB, domainGame.GameTypeTeamPlay, domainGame.VariantGross)
	if err != nil {
		t.Fatalf("NewGame failed: %v", err)
	}
	if err := gameRepo.CreateGame(ctx, g); err != nil {
		t.Fatalf("CreateGame failed: %v", err)
	}

	// Score hole 1: reuse your hand-verified worked example (par 4).
	inputs := []domainGame.PlayerScoreInput{
		{PlayerID: teamA[0].ID, Gross: 3, TeamID: "A"}, // birdie
		{PlayerID: teamA[1].ID, Gross: 8, TeamID: "A"},
		{PlayerID: teamB[0].ID, Gross: 4, TeamID: "B"},
		{PlayerID: teamB[1].ID, Gross: 6, TeamID: "B"},
	}

	if err := g.SetHoleScore(1, inputs); err != nil {
		t.Fatalf("SetHoleScore failed: %v", err)
	}
	if err := gameRepo.SaveHoleResult(ctx, g, 1); err != nil {
		t.Fatalf("SaveHoleResult failed: %v", err)
	}

	// Reload from scratch — this is the real test: does LoadGame's
	// findHoleResults/findHoleResultScores correctly reconstruct
	// everything we just wrote?
	loaded, err := gameRepo.LoadGame(ctx, "test-game-hole-result")
	if err != nil {
		t.Fatalf("LoadGame failed: %v", err)
	}

	if loaded.CurrentHole != 2 {
		t.Errorf("loaded.CurrentHole = %d, want 2 (auto-advanced after scoring hole 1)", loaded.CurrentHole)
	}

	result, ok := loaded.HoleResults[1]
	if !ok {
		t.Fatal("loaded game has no HoleResults[1] — hole result was not persisted or not reloaded")
	}

	if result.PointsA != 2 || result.PointsB != 1 {
		t.Errorf("reloaded PointsA/B = %d/%d, want 2/1", result.PointsA, result.PointsB)
	}
	if result.LowScoreWinnerTeam != "A" {
		t.Errorf("reloaded LowScoreWinnerTeam = %q, want A", result.LowScoreWinnerTeam)
	}
	if result.TeamTotalWinnerTeam != "B" {
		t.Errorf("reloaded TeamTotalWinnerTeam = %q, want B", result.TeamTotalWinnerTeam)
	}
	if len(result.Scores) != 4 {
		t.Fatalf("reloaded has %d player scores, want 4", len(result.Scores))
	}
	if len(result.GrossBonuses) != 1 {
		t.Fatalf("reloaded has %d gross bonuses, want 1 (birdie)", len(result.GrossBonuses))
	}

	// Confirm the specific birdie scorer's data round-tripped correctly.
	var p1Score *domainGame.PlayerHoleResult
	for i := range result.Scores {
		if result.Scores[i].PlayerID == teamA[0].ID {
			p1Score = &result.Scores[i]
		}
	}
	if p1Score == nil {
		t.Fatal("expected a score entry for teamA[0]")
	}
	if p1Score.Gross != 3 || p1Score.GrossBonus != 1 {
		t.Errorf("reloaded teamA[0] Gross/GrossBonus = %d/%d, want 3/1", p1Score.Gross, p1Score.GrossBonus)
	}

	// Loaded match score should reflect the same hole result (StartingLead
	// is 0 here since all handicaps are equal).
	wantScore := domainGame.MatchScore{TeamA: 1, TeamB: 0}
	if loaded.MatchScore != wantScore {
		t.Errorf("loaded.MatchScore = %+v, want %+v", loaded.MatchScore, wantScore)
	}
}
