package opengolfapi

import "golf-game-kaffip/internal/domain/course"

// MapCourseToDomain converts the external API's wire-format course
// response into the domain course model.
func MapCourseToDomain(res *CourseResponse) *course.Course {
	return &course.Course{
		ID:         res.ID,
		Name:       res.CourseName,
		ClubName:   res.ClubName,
		City:       res.City,
		State:      res.State,
		Lat:        res.Lat,
		Lng:        res.Lng,
		Type:       res.Type,
		Par:        res.Par,
		Holes:      res.Holes,
		Yardage:    res.Yardage,
		Timezone:   res.Timezone,
		YearBuilt:  res.YearBuilt,
		Phone:      safeString(res.Phone),
		Website:    safeString(res.Website),
		Address:    safeString(res.Address),
		PostalCode: safeString(res.PostalCode),
		Tees:       mapTees(res.Tees),
		HolesData:  mapHoles(res.HolesData),
	}
}

func safeString(v any) string {
	s, _ := v.(string)
	return s
}

func mapTees(tees []Tee) []course.Tee {
	out := make([]course.Tee, 0, len(tees))
	for _, t := range tees {
		out = append(out, course.Tee{
			Key:          t.TeeKey,
			Name:         t.TeeName,
			Color:        t.TeeColor,
			Gender:       t.Gender,
			CourseRating: t.CourseRating,
			Slope:        t.Slope,
			Par:          t.Par,
			Yardage:      t.Yardage,
		})
	}
	return out
}

func mapHoles(data []HoleData) []course.Hole {
	holes := make([]course.Hole, 0, len(data))
	for _, h := range data {
		holes = append(holes, course.Hole{
			Number:        h.Number,
			Par:           h.Par,
			HandicapIndex: h.HandicapIndex,
			Yardages:      h.Yardages,
		})
	}
	return holes
}
