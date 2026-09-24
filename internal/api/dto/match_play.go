package dto

type CreateMatchPlayRequest struct {
	CourseID string `json:"course_id"`
	Variant  string `json:"variant"`
	PlayerA  int64  `json:"player_a"`
	PlayerB  int64  `json:"player_b"`
}
