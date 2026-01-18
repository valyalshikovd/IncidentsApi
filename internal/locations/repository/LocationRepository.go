package location

import (
	"RedColarTest/internal/common"
	domain "RedColarTest/internal/locations/domain"
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	db *pgxpool.Pool
}

func NewLocationRepo(db *pgxpool.Pool) *Repo {
	return &Repo{db: db}
}

func (r *Repo) SaveCheck(ctx context.Context, in domain.LocationCheck) (int64, *common.Error) {
	const q = `
insert into location_checks (user_id, latitude, longitude, has_danger)
values ($1, $2, $3, $4)
returning id;
`
	var id int64
	if err := r.db.QueryRow(ctx, q, in.UserID, in.Latitude, in.Longitude, in.HasDanger).Scan(&id); err != nil {
		return 0, common.NewError(common.CodeIternalErr, err.Error())
	}
	return id, nil
}

func (r *Repo) CountUniqueUsersSince(ctx context.Context, since time.Time) (int64, *common.Error) {
	const q = `select count(distinct user_id) from location_checks where created_at >= $1;`
	var count int64
	if err := r.db.QueryRow(ctx, q, since).Scan(&count); err != nil {
		return 0, common.NewError(common.CodeIternalErr, err.Error())
	}
	return count, nil
}
