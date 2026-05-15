package repository

import (
	"context"

	"github.com/yourteam/cinema-system/movie-service/internal/domain"
)

func (r *hallRepo) List(ctx context.Context) ([]domain.Hall, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, rows, cols FROM halls ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var halls []domain.Hall
	for rows.Next() {
		var h domain.Hall
		if err := rows.Scan(&h.ID, &h.Name, &h.Rows, &h.Cols); err != nil {
			return nil, err
		}
		halls = append(halls, h)
	}
	if halls == nil {
		halls = []domain.Hall{}
	}
	return halls, nil
}
