package cup

import "testing"

func TestNewCup_RequiresBothSides(t *testing.T) {
	players := map[int64]Side{1: SideA, 2: SideA} // no one on side B
	_, err := NewCup("cup_1", "Test Cup", players, []string{"game_1"})
	if err == nil {
		t.Fatal("expected error when no player is on side B, got nil")
	}
}

func TestNewCup_Valid(t *testing.T) {
	players := map[int64]Side{1: SideA, 2: SideB}
	c, err := NewCup("cup_1", "Test Cup", players, []string{"game_1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.ID != "cup_1" || len(c.Players) != 2 || len(c.MatchIDs) != 1 {
		t.Errorf("unexpected Cup: %+v", c)
	}
}
