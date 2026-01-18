package repository

import (
	"RedColarTest/internal/common"
	"RedColarTest/internal/incident/domain"
	"context"
)

type IncidentRepository interface {
	Create(ctx context.Context, in domain.Incident) (domain.Incident, *common.Error)

	GetByID(ctx context.Context, id int64) (domain.Incident, *common.Error)

	List(ctx context.Context, limit, offset int, onlyActive bool) ([]domain.Incident, int64, *common.Error)

	ListActive(ctx context.Context) ([]domain.Incident, *common.Error)

	Update(ctx context.Context, id int64, in domain.Incident) (domain.Incident, *common.Error)

	Deactivate(ctx context.Context, id int64) (domain.Incident, *common.Error)
}
