package game

import "time"

type GameRow struct {
	ID           string
	CourseID     string
	CourseName   string
	Variant      string
	StartingLead int
	CurrentHole  int
	MatchTeamA   int
	MatchTeamB   int
	CreatedAt    time.Time
	UpdatedAt    time.Time
	FinishedAt   *time.Time
}

type CourseHoleRow struct {
	HoleNumber    int
	Par           int
	HandicapIndex int
}

type HoleResultRow struct {
	ID                  int64
	GameID              string
	HoleNumber          int
	PointsA             int
	PointsB             int
	LowScoreWinnerTeam  *string
	TeamTotalWinnerTeam *string
}

type HoleResultScoreRow struct {
	PlayerID   int64
	Gross      int
	Net        int
	Strokes    int
	GrossBonus int
}

type GameSummaryRow struct {
	ID          string
	CourseID    string
	CourseName  string
	Variant     string
	CurrentHole int
	TotalHoles  int
	MatchTeamA  int
	MatchTeamB  int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	FinishedAt  *time.Time
}
