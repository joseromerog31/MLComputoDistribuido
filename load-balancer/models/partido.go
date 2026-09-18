package models

import "database/sql"

type PartidoFeature struct {
	ID                int64   `json:"id"`
	HomeAvgGoalsLast5 float64 `json:"home_avg_goals_last5"`
}

type PartidoModel struct {
	DB *sql.DB
}

func (m *PartidoModel) GetBatch(
	limit int,
) ([]PartidoFeature, error) {

	query := `
		SELECT
			id,
			home_avg_goals_last5
		FROM ml_features
		WHERE home_avg_goals_last5 IS NOT NULL
		ORDER BY date
		LIMIT $1
	`

	rows, err := m.DB.Query(
		query,
		limit,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var partidos []PartidoFeature

	for rows.Next() {

		var partido PartidoFeature

		err := rows.Scan(
			&partido.ID,
			&partido.HomeAvgGoalsLast5,
		)

		if err != nil {
			return nil, err
		}

		partidos = append(
			partidos,
			partido,
		)
	}

	return partidos, rows.Err()
}
