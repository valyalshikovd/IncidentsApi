package services

import (
	"RedColarTest/internal/common"
	"RedColarTest/internal/incident/domain"
	"RedColarTest/internal/incident/repository"
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type IncidentService struct {
	repo     repository.IncidentRepository
	cache    *redis.Client
	cacheKey string
}

func NewIncidentService(repo repository.IncidentRepository, cache *redis.Client) *IncidentService {
	return &IncidentService{repo: repo, cache: cache, cacheKey: "cache:active_incidents"}
}

func (s *IncidentService) Create(ctx context.Context, in domain.Incident) (*domain.Incident, *common.Error) {
	if err := validateIncident(in); err != nil {
		return &domain.Incident{}, common.NewError(common.CodeNotValid, err.Error())
	}

	incident, err := s.repo.Create(ctx, in)

	if err != nil {
		return nil, err
	}
	s.invalidateCache(ctx)
	return &incident, nil
}

func (s *IncidentService) GetByID(ctx context.Context, id int64) (domain.Incident, *common.Error) {
	if id <= 0 {
		return domain.Incident{}, common.NewError(common.CodeNotValid, fmt.Sprintf("Incident with id %d not found", id))
	}
	incident, err := s.repo.GetByID(ctx, id)
	return incident, err
}

func (s *IncidentService) List(ctx context.Context, page, pageSize int, onlyActive bool) ([]domain.Incident, int64, int, int, *common.Error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}
	offset := (page - 1) * pageSize

	items, total, err := s.repo.List(ctx, pageSize, offset, onlyActive)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	return items, total, page, pageSize, nil
}

func (s *IncidentService) Update(ctx context.Context, id int64, in domain.Incident) (domain.Incident, *common.Error) {
	if id <= 0 {
		return domain.Incident{}, common.NewError(common.CodeNotValid, fmt.Sprintf("Incident with id %d not found", id))
	}
	if err := validateIncident(in); err != nil {
		return domain.Incident{}, common.NewError(common.CodeNotValid, err.Error())
	}
	out, err := s.repo.Update(ctx, id, in)
	if err == nil {
		s.invalidateCache(ctx)
	}
	return out, err
}

func (s *IncidentService) Deactivate(ctx context.Context, id int64) (domain.Incident, *common.Error) {
	if id <= 0 {
		return domain.Incident{}, common.NewError(common.CodeNotValid, "Incident with id < 0 invalid")
	}
	out, err := s.repo.Deactivate(ctx, id)
	if err == nil {
		s.invalidateCache(ctx)
	}
	return out, err
}

func validateIncident(in domain.Incident) error {
	if in.Title == "" {
		return fmt.Errorf("title is required")
	}
	if in.Latitude < -90 || in.Latitude > 90 {
		return fmt.Errorf("latitude out of range")
	}
	if in.Longitude < -180 || in.Longitude > 180 {
		return fmt.Errorf("longitude out of range")
	}
	if in.DangerRadiusM <= 0 {
		return fmt.Errorf("danger_radius_m must be > 0")
	}
	return nil
}

func (s *IncidentService) invalidateCache(ctx context.Context) {
	if s.cache == nil {
		return
	}
	_ = s.cache.Del(ctx, s.cacheKey).Err()
}
