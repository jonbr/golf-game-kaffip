package player_test

import (
	"context"
	"errors"
	"golf-game-kaffip/internal/testutil"
	"testing"

	domainPlayer "golf-game-kaffip/internal/domain/player"
	playerrepo "golf-game-kaffip/internal/infrastructure/postgres/player"
)

func TestPlayerRepository_CreateAndFindWeithEmail(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	ctx := context.Background()
	repo := playerrepo.NewRepository(pool)

	p := &domainPlayer.Player{
		Name:     "Test Player",
		Email:    "test@example.com",
		Handicap: 12.5,
	}

	if err := repo.Create(ctx, p); err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if p.ID == 0 {
		t.Fatalf("expected Crete to populate ID")
	}

	loaded, err := repo.FindByID(ctx, p.ID, false)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}

	if loaded.Email != "test@example.com" {
		t.Errorf("loaded.Email = %q, want test@example.com", loaded.Email)
	}
}

func TestPlayerRepository_CreateWithDuplicateEmailFails(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	ctx := context.Background()
	repo := playerrepo.NewRepository(pool)

	email := "duplicate@example.com"

	first := &domainPlayer.Player{Name: "First", Email: email, Handicap: 10}
	if err := repo.Create(ctx, first); err != nil {
		t.Fatalf("first Create failed: %v", err)
	}

	second := &domainPlayer.Player{Name: "Second", Email: email, Handicap: 15}
	err := repo.Create(ctx, second)
	if err == nil {
		t.Fatal("expected error for duplicate email, got nil")
	}
	if !errors.Is(err, domainPlayer.ErrEmailAlreadyExists) {
		t.Errorf("expected ErrEmailAlreadyExists, got: %v", err)
	}
}

func TestPlayerRepository_FindAllReturnsEmailForEveryPlayer(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	ctx := context.Background()
	repo := playerrepo.NewRepository(pool)

	players := []*domainPlayer.Player{
		{Name: "Alice", Email: "alice@example.com", Handicap: 5},
		{Name: "Bob", Email: "bob@example.com", Handicap: 15},
	}
	for _, p := range players {
		if err := repo.Create(ctx, p); err != nil {
			t.Fatalf("Create failed for %s: %v", p.Name, err)
		}
	}

	all, err := repo.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll failed: %v", err)
	}

	if len(all) != 2 {
		t.Fatalf("FindAll returned %d players, want 2", len(all))
	}
	for _, p := range all {
		if p.Email == "" {
			t.Errorf("player %d (%s) has empty email, want a real value", p.ID, p.Name)
		}
	}
}
