package models

import "database/sql"

// Info mínima para que el ML jale
type PartidoFeature struct {
	ID                int64   `json:"id"`
	HomeAvgGoalsLast5 float64 `json:"home_avg_goals_last5"`
}

// Recupera la info. de la vista para ML de Postgres
type PartidoModel struct {
	DB *sql.DB
}

// Batch cronológico de registros
// El feature engineering se hace en Posgress
// ¿Bueno o malo? No lo se pero el coordinator recibe data ya lista
func (m *PartidoModel) GetBatch(
	limit int,
) ([]PartidoFeature, error) {

	query := `
		SELECT
			id,
			home_avg_goals_last5
		FROM ml_features
		WHERE home_avg_goals_last5 IS NOT NULL
		ORDER BY date, id
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

	// COnvierte cada registro en una estructura más ligera
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
