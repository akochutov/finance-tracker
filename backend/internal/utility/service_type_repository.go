package utility

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ServiceTypeRepository struct {
	db *pgxpool.Pool
}

func NewServiceTypeRepository(db *pgxpool.Pool) *ServiceTypeRepository {
	return &ServiceTypeRepository{db: db}
}

func (r *ServiceTypeRepository) List(ctx context.Context) ([]ServiceType, error) {
	const q = `SELECT code, name, unit FROM services ORDER BY name`

	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list services: %w", err)
	}
	defer rows.Close()

	services := make([]ServiceType, 0)
	for rows.Next() {
		var s ServiceType
		err := rows.Scan(&s.Code, &s.Name, &s.Unit)
		if err != nil {
			return nil, fmt.Errorf("scan services: %w", err)
		}
		services = append(services, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate services: %w", err)
	}

	return services, nil
}
