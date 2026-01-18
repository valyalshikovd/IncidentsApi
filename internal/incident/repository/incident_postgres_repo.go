package repository

import (
	"RedColarTest/internal/common"
	"RedColarTest/internal/incident/domain"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type IncidentRepo struct {
	db *pgxpool.Pool
}

func NewIncidentRepo(db *pgxpool.Pool) *IncidentRepo {
	return &IncidentRepo{db: db}
}

func (r *IncidentRepo) Create(ctx context.Context, in domain.Incident) (domain.Incident, *common.Error) {
	const q = `
insert into incidents (title, description, latitude, longitude, danger_radius_m, is_active)
values ($1, $2, $3, $4, $5, $6)
returning id, title, description, latitude, longitude, danger_radius_m, is_active, created_at, updated_at;
`
	var out domain.Incident
	err := r.db.QueryRow(ctx, q,
		in.Title,
		in.Description,
		in.Latitude,
		in.Longitude,
		in.DangerRadiusM,
		in.IsActive,
	).Scan(
		&out.ID,
		&out.Title,
		&out.Description,
		&out.Latitude,
		&out.Longitude,
		&out.DangerRadiusM,
		&out.IsActive,
		&out.CreatedAt,
		&out.UpdatedAt,
	)
	if err != nil {
		return domain.Incident{}, common.NewError(common.CodeIternalErr, err.Error())
	}
	return out, nil
}

func (r *IncidentRepo) GetByID(ctx context.Context, id int64) (domain.Incident, *common.Error) {
	const q = `
select id, title, description, latitude, longitude, danger_radius_m, is_active, created_at, updated_at
from incidents
where id = $1;
`
	var out domain.Incident
	err := r.db.QueryRow(ctx, q, id).Scan(
		&out.ID,
		&out.Title,
		&out.Description,
		&out.Latitude,
		&out.Longitude,
		&out.DangerRadiusM,
		&out.IsActive,
		&out.CreatedAt,
		&out.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Incident{}, common.NewError(common.CodeNotFound, err.Error())
	}
	if err != nil {
		return domain.Incident{}, common.NewError(common.CodeIternalErr, err.Error())
	}
	return out, nil
}

func (r *IncidentRepo) List(ctx context.Context, limit, offset int, onlyActive bool) ([]domain.Incident, int64, *common.Error) {
	totalQ := `select count(1) from incidents where ($1 = false) or (is_active = true);`
	var total int64
	if err := r.db.QueryRow(ctx, totalQ, onlyActive).Scan(&total); err != nil {
		return nil, 0, common.NewError(common.CodeIternalErr, err.Error())
	}

	const q = `
    select id, title, description, latitude, longitude, danger_radius_m, is_active, created_at, updated_at
    from incidents
    where ($1 = false) or (is_active = true)
    order by id desc
    limit $2 offset $3;
    `
	rows, err := r.db.Query(ctx, q, onlyActive, limit, offset)
	if err != nil {
		return nil, 0, common.NewError(common.CodeIternalErr, err.Error())
	}
	defer rows.Close()

	items := make([]domain.Incident, 0, limit)
	for rows.Next() {
		var it domain.Incident
		if err := rows.Scan(
			&it.ID,
			&it.Title,
			&it.Description,
			&it.Latitude,
			&it.Longitude,
			&it.DangerRadiusM,
			&it.IsActive,
			&it.CreatedAt,
			&it.UpdatedAt,
		); err != nil {
			return nil, 0, common.NewError(common.CodeIternalErr, err.Error())
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, common.NewError(common.CodeIternalErr, err.Error())
	}
	return items, total, nil
}

func (r *IncidentRepo) ListActive(ctx context.Context) ([]domain.Incident, *common.Error) {
	const q = `
    select id, title, description, latitude, longitude, danger_radius_m, is_active, created_at, updated_at
    from incidents
    where is_active = true
    order by id desc;
    `
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, common.NewError(common.CodeIternalErr, err.Error())
	}
	defer rows.Close()

	items := make([]domain.Incident, 0)
	for rows.Next() {
		var it domain.Incident
		if err := rows.Scan(
			&it.ID,
			&it.Title,
			&it.Description,
			&it.Latitude,
			&it.Longitude,
			&it.DangerRadiusM,
			&it.IsActive,
			&it.CreatedAt,
			&it.UpdatedAt,
		); err != nil {
			return nil, common.NewError(common.CodeIternalErr, err.Error())
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, common.NewError(common.CodeIternalErr, err.Error())
	}
	return items, nil
}

func (r *IncidentRepo) Update(ctx context.Context, id int64, in domain.Incident) (domain.Incident, *common.Error) {
	const q = `
        update incidents
        set title = $2,
        description = $3,
        latitude = $4,
        longitude = $5,
        danger_radius_m = $6,
        is_active = $7,
        updated_at = now()
        where id = $1
        returning id, title, description, latitude, longitude, danger_radius_m, is_active, created_at, updated_at;
`
	var out domain.Incident
	err := r.db.QueryRow(ctx, q,
		id,
		in.Title,
		in.Description,
		in.Latitude,
		in.Longitude,
		in.DangerRadiusM,
		in.IsActive,
	).Scan(
		&out.ID,
		&out.Title,
		&out.Description,
		&out.Latitude,
		&out.Longitude,
		&out.DangerRadiusM,
		&out.IsActive,
		&out.CreatedAt,
		&out.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Incident{}, common.NewError(common.CodeNotFound, err.Error())
	}
	if err != nil {
		return domain.Incident{}, common.NewError(common.CodeIternalErr, err.Error())
	}
	return out, nil
}

func (r *IncidentRepo) Deactivate(ctx context.Context, id int64) (domain.Incident, *common.Error) {
	const q = `
    update incidents
    set is_active = false,
        deactivated_at = now(),
        updated_at = now()
    where id = $1 and is_active = true
    returning id, title, description, latitude, longitude, danger_radius_m, is_active, created_at, updated_at;
`
	var out domain.Incident
	err := r.db.QueryRow(ctx, q, id).Scan(
		&out.ID,
		&out.Title,
		&out.Description,
		&out.Latitude,
		&out.Longitude,
		&out.DangerRadiusM,
		&out.IsActive,
		&out.CreatedAt,
		&out.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Incident{}, common.NewError(common.CodeNotFound, err.Error())
	}
	if err != nil {
		return domain.Incident{}, common.NewError(common.CodeIternalErr, err.Error())
	}
	return out, nil
}
