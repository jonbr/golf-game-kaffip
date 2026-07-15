package application

import (
	"context"
	"golf-game-kaffip/internal/domain/course"
	"golf-game-kaffip/internal/infrastructure/external/opengolfapi"
)

type ExternalCourseService struct {
	api opengolfapi.ClientInterface
}

func NewExternalCourseService(api opengolfapi.ClientInterface) *ExternalCourseService {
	return &ExternalCourseService{api: api}
}

func (s *ExternalCourseService) GetExternalCourse(ctx context.Context, id string) (*course.Course, error) {
	res, err := s.api.GetCourse(ctx, id)
	if err != nil {
		return nil, err
	}
	return opengolfapi.MapCourseToDomain(res), nil
}
