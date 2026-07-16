package game

import (
	"testing"
)

// newTestGame builds a minimal 3-hole VariantGross game for testing
// SetHoleScore's orchestration logic (advance, reject, replay).
func newTestGame(t *testing.T) *Game {
	t.Helper()

	/*c := &course.Course{
		ID: "test-course",
		HolesData: []course.Hole{
			{Number: 1, Par: 4, HandicapIndex: 9},
			{Number: 2, Par: 3, HandicapIndex: 11},
			{Number: 3, Par: 5, HandicapIndex: 1},
		},
	}*/

	//teamA := []*player.Player{{ID: 1, Handicap: 0}, {ID: 2, Handicap: 0}}
	//teamB := []*player.Player{{ID: 3, Handicap: 0}, {ID: 4, Handicap: 0}}

	//g, err := NewGame("test-game", c, teamA, teamB, VariantGross)
	/*if err != nil {
		t.Fatalf("NewGame failed: %v", err)
	}*/
	return nil
}

func scores(p1, p2, p3, p4 int) []PlayerScoreInput {
	return []PlayerScoreInput{
		{PlayerID: 1, Gross: p1, TeamID: "A"},
		{PlayerID: 2, Gross: p2, TeamID: "A"},
		{PlayerID: 3, Gross: p3, TeamID: "B"},
		{PlayerID: 4, Gross: p4, TeamID: "B"},
	}
}

func TestSetHoleScore_FreshSubmissionAdvancesCurrentHole(t *testing.T) {
	g := newTestGame(t)

	if g.CurrentHole != 1 {
		t.Fatalf("expected new game to start at hole 1, got %d", g.CurrentHole)
	}

	if err := g.SetHoleScore(1, scores(3, 8, 4, 6)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if g.CurrentHole != 2 {
		t.Errorf("CurrentHole = %d, want 2 after scoring hole 1", g.CurrentHole)
	}
}

func TestSetHoleScore_RejectsScoringAheadOfCurrentHole(t *testing.T) {
	g := newTestGame(t)

	err := g.SetHoleScore(2, scores(4, 4, 4, 4))
	if err == nil {
		t.Fatal("expected error scoring hole 2 while CurrentHole is 1, got nil")
	}
}

func TestSetHoleScore_RejectsHoleZeroOrNegative(t *testing.T) {
	g := newTestGame(t)

	if err := g.SetHoleScore(0, scores(4, 4, 4, 4)); err == nil {
		t.Error("expected error for hole 0, got nil")
	}
	if err := g.SetHoleScore(-1, scores(4, 4, 4, 4)); err == nil {
		t.Error("expected error for negative hole, got nil")
	}
}

func TestSetHoleScore_CorrectionDoesNotAdvanceCurrentHole(t *testing.T) {
	g := newTestGame(t)

	// Score holes 1, 2, 3 in order — CurrentHole should reach 3 (capped, no hole 4 exists).
	mustScore(t, g, 1, scores(3, 8, 4, 6))
	mustScore(t, g, 2, scores(4, 4, 4, 4))
	mustScore(t, g, 3, scores(5, 5, 5, 5))

	if g.CurrentHole != 3 {
		t.Fatalf("CurrentHole = %d, want 3 (capped at course's total holes)", g.CurrentHole)
	}

	// Now correct hole 1. CurrentHole must NOT move, since this is a
	// correction to an earlier hole, not a fresh submission.
	if err := g.SetHoleScore(1, scores(4, 4, 4, 4)); err != nil {
		t.Fatalf("unexpected error correcting hole 1: %v", err)
	}

	if g.CurrentHole != 3 {
		t.Errorf("CurrentHole = %d, want unchanged at 3 after correcting hole 1", g.CurrentHole)
	}
}

func TestSetHoleScore_CurrentHoleNeverExceedsTotalHoles(t *testing.T) {
	g := newTestGame(t) // 3-hole course

	mustScore(t, g, 1, scores(4, 4, 4, 4))
	mustScore(t, g, 2, scores(4, 4, 4, 4))
	mustScore(t, g, 3, scores(4, 4, 4, 4))

	if g.CurrentHole != 3 {
		t.Errorf("CurrentHole = %d, want capped at 3 (course has only 3 holes)", g.CurrentHole)
	}

	// Re-scoring hole 3 again (a correction, since CurrentHole is already
	// at 3 and this equals it — edge case: is this "fresh" or "correction"?)
	// Per design: holeNumber == CurrentHole is treated as fresh, but since
	// CurrentHole is already capped at the max, it should stay capped.
	if err := g.SetHoleScore(3, scores(5, 5, 5, 5)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.CurrentHole != 3 {
		t.Errorf("CurrentHole = %d, want still capped at 3", g.CurrentHole)
	}
}

func TestSetHoleScore_ReplayRecomputesMatchScoreAfterCorrection(t *testing.T) {
	g := newTestGame(t)

	// Hole 1 (par 4): A=3(birdie),8  B=4,6 -> low=A(1), total=B(11v10? wait recompute)
	mustScore(t, g, 1, scores(3, 8, 4, 6))
	// Hole 2 (par 3): even score, no winner either category
	mustScore(t, g, 2, scores(4, 4, 4, 4))
	// Hole 3 (par 5): team B wins low + total clearly
	mustScore(t, g, 3, scores(6, 6, 4, 5))

	leadBeforeCorrection := g.MatchScore

	// Correct hole 1 to a tie (no points for anyone that hole).
	if err := g.SetHoleScore(1, scores(4, 4, 4, 4)); err != nil {
		t.Fatalf("unexpected error correcting hole 1: %v", err)
	}

	if g.MatchScore == leadBeforeCorrection {
		t.Fatalf("expected MatchScore to change after correcting hole 1, but it stayed %+v", g.MatchScore)
	}

	// Manually recompute what the score SHOULD be after the correction,
	// to assert the replay is exactly right, not just "different".
	wantLead := g.StartingLead
	for h := 1; h <= g.CurrentHole; h++ {
		r, ok := g.HoleResults[h]
		if !ok {
			continue
		}
		wantLead += r.PointsA - r.PointsB
	}
	wantScore := matchScoreFromLead(wantLead)

	if g.MatchScore != wantScore {
		t.Errorf("MatchScore = %+v, want %+v (recomputed from stored hole results)", g.MatchScore, wantScore)
	}
}

func TestSetHoleScore_UnknownPlayerRejected(t *testing.T) {
	g := newTestGame(t)

	badScores := []PlayerScoreInput{
		{PlayerID: 999, Gross: 4, TeamID: "A"},
		{PlayerID: 2, Gross: 4, TeamID: "A"},
		{PlayerID: 3, Gross: 4, TeamID: "B"},
		{PlayerID: 4, Gross: 4, TeamID: "B"},
	}

	if err := g.SetHoleScore(1, badScores); err == nil {
		t.Error("expected error for unknown player id, got nil")
	}
}

func mustScore(t *testing.T, g *Game, hole int, s []PlayerScoreInput) {
	t.Helper()
	if err := g.SetHoleScore(hole, s); err != nil {
		t.Fatalf("SetHoleScore(%d) failed: %v", hole, err)
	}
}
