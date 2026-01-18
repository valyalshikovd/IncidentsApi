package location

import (
	"RedColarTest/internal/common"
	incdomain "RedColarTest/internal/incident/domain"
	increpo "RedColarTest/internal/incident/repository"
	domain "RedColarTest/internal/locations/domain"
	locationrepo "RedColarTest/internal/locations/repository"
	"RedColarTest/internal/webhook"
	"context"
	"encoding/json"
	"errors"
	"math"
	"sort"
	"time"

	"github.com/redis/go-redis/v9"
)

type Service struct {
	repo      *locationrepo.Repo
	incRepo   increpo.IncidentRepository
	cache     *redis.Client
	cacheTTL  time.Duration
	webhookQ  *webhook.Queue
	cacheKey  string
	cacheLive bool
}

func NewLocationService(
	repo *locationrepo.Repo,
	incRepo increpo.IncidentRepository,
	cache *redis.Client,
	cacheTTL time.Duration,
	webhookQ *webhook.Queue,
) *Service {
	return &Service{
		repo:      repo,
		incRepo:   incRepo,
		cache:     cache,
		cacheTTL:  cacheTTL,
		webhookQ:  webhookQ,
		cacheKey:  "cache:active_incidents",
		cacheLive: cache != nil && cacheTTL > 0,
	}
}

func (s *Service) CheckLocation(ctx context.Context, userID string, lat, lon float64) (domain.CheckResult, *common.Error) {
	if s.incRepo == nil {
		return domain.CheckResult{}, common.NewError(common.CodeIternalErr, "incident repo is not initialized")
	}
	if err := validateLocationInput(userID, lat, lon); err != nil {
		return domain.CheckResult{}, err
	}

	incidents, err := s.getActiveIncidents(ctx)
	if err != nil {
		return domain.CheckResult{}, err
	}

	matches := make([]domain.IncidentDistance, 0)
	for _, inc := range incidents {
		dist := haversineMeters(lat, lon, inc.Latitude, inc.Longitude)
		if dist <= float64(inc.DangerRadiusM) {
			matches = append(matches, domain.IncidentDistance{
				IncidentID:    inc.ID,
				Title:         inc.Title,
				Description:   inc.Description,
				Latitude:      inc.Latitude,
				Longitude:     inc.Longitude,
				DangerRadiusM: inc.DangerRadiusM,
				DistanceM:     dist,
			})
		}
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].DistanceM < matches[j].DistanceM
	})

	return domain.CheckResult{
		Dangerous: len(matches) > 0,
		Incidents: matches,
	}, nil
}

func (s *Service) RecordCheck(ctx context.Context, userID string, lat, lon float64, incidents []domain.IncidentDistance) (int64, *common.Error) {
	if s.repo == nil {
		return 0, common.NewError(common.CodeIternalErr, "location repo is not initialized")
	}
	if err := validateLocationInput(userID, lat, lon); err != nil {
		return 0, err
	}
	checkID, err := s.repo.SaveCheck(ctx, domain.LocationCheck{
		UserID:    userID,
		Latitude:  lat,
		Longitude: lon,
		HasDanger: len(incidents) > 0,
	})
	if err != nil {
		return 0, err
	}

	if len(incidents) > 0 && s.webhookQ != nil {
		payload := webhook.Payload{
			CheckID:   checkID,
			UserID:    userID,
			Latitude:  lat,
			Longitude: lon,
			Incidents: mapWebhookIncidents(incidents),
			CreatedAt: time.Now().UTC(),
		}
		_ = s.webhookQ.Enqueue(ctx, payload)
	}

	return checkID, nil
}

func (s *Service) Stats(ctx context.Context, windowMinutes int) (int64, *common.Error) {
	if s.repo == nil {
		return 0, common.NewError(common.CodeIternalErr, "location repo is not initialized")
	}
	if windowMinutes <= 0 {
		return 0, common.NewError(common.CodeNotValid, "stats window must be > 0")
	}
	since := time.Now().Add(-time.Duration(windowMinutes) * time.Minute)
	return s.repo.CountUniqueUsersSince(ctx, since)
}

func (s *Service) getActiveIncidents(ctx context.Context) ([]incdomain.Incident, *common.Error) {
	if s.incRepo == nil {
		return nil, common.NewError(common.CodeIternalErr, "incident repo is not initialized")
	}
	if !s.cacheLive {
		return s.incRepo.ListActive(ctx)
	}

	raw, err := s.cache.Get(ctx, s.cacheKey).Result()
	if err == nil {
		var cached []incdomain.Incident
		if err := json.Unmarshal([]byte(raw), &cached); err == nil {
			return cached, nil
		}
		_ = s.cache.Del(ctx, s.cacheKey).Err()
	} else if !errors.Is(err, redis.Nil) {
		_ = s.cache.Del(ctx, s.cacheKey).Err()
	}

	incidents, repoErr := s.incRepo.ListActive(ctx)
	if repoErr != nil {
		return nil, repoErr
	}

	if payload, err := json.Marshal(incidents); err == nil {
		_ = s.cache.Set(ctx, s.cacheKey, payload, s.cacheTTL).Err()
	}

	return incidents, nil
}

func mapWebhookIncidents(incidents []domain.IncidentDistance) []webhook.PayloadIncident {
	out := make([]webhook.PayloadIncident, 0, len(incidents))
	for _, inc := range incidents {
		out = append(out, webhook.PayloadIncident{
			ID:            inc.IncidentID,
			Title:         inc.Title,
			Description:   inc.Description,
			Latitude:      inc.Latitude,
			Longitude:     inc.Longitude,
			DangerRadiusM: inc.DangerRadiusM,
			DistanceM:     inc.DistanceM,
		})
	}
	return out
}

func validateLocationInput(userID string, lat, lon float64) *common.Error {
	if userID == "" {
		return common.NewError(common.CodeNotValid, "user_id is required")
	}
	if lat < -90 || lat > 90 {
		return common.NewError(common.CodeNotValid, "latitude out of range")
	}
	if lon < -180 || lon > 180 {
		return common.NewError(common.CodeNotValid, "longitude out of range")
	}
	return nil
}

func haversineMeters(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371000.0
	rad := func(d float64) float64 { return d * math.Pi / 180 }
	dLat := rad(lat2 - lat1)
	dLon := rad(lon2 - lon1)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(rad(lat1))*math.Cos(rad(lat2))*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadius * c
}
