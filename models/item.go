package models

import "database/sql"

type Partido struct {
	ID                   string `json:"id"`
	EquipoLocal          string `json:"equipo_local"`
	EquipoVisitante      string `json:"equipo_visitante"`
	GolesEquipoLocal     int    `json:"goles_equipo_local"`
	GolesEquipoVisitante int    `json:"goles_equipo_visitante"`
	Jornada              int    `json:"jornada"`
	Estadio              string `json:"estadio"`
}

type PartidoModel struct {
	DB *sql.DB
}

// INSERTAR
func (m *PartidoModel) Create(partido *Partido) error {

	query := `
		INSERT INTO partidos (
			id,
			equipo_local,
			equipo_visitante,
			goles_equipo_local,
			goles_equipo_visitante,
			jornada,
			estadio
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := m.DB.Exec(
		query,
		partido.ID,
		partido.EquipoLocal,
		partido.EquipoVisitante,
		partido.GolesEquipoLocal,
		partido.GolesEquipoVisitante,
		partido.Jornada,
		partido.Estadio,
	)

	return err
}

// LEER TODOS
func (m *PartidoModel) GetAll() ([]Partido, error) {

	query := `
		SELECT
			id,
			equipo_local,
			equipo_visitante,
			goles_equipo_local,
			goles_equipo_visitante,
			jornada,
			estadio
		FROM partidos
		ORDER BY jornada, id
	`

	rows, err := m.DB.Query(query)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	partidos := []Partido{}

	for rows.Next() {

		var partido Partido

		err := rows.Scan(
			&partido.ID,
			&partido.EquipoLocal,
			&partido.EquipoVisitante,
			&partido.GolesEquipoLocal,
			&partido.GolesEquipoVisitante,
			&partido.Jornada,
			&partido.Estadio,
		)

		if err != nil {
			return nil, err
		}

		partidos = append(partidos, partido)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return partidos, nil
}

// LEER UNO
func (m *PartidoModel) GetByID(id string) (*Partido, error) {

	partido := &Partido{}

	query := `
		SELECT
			id,
			equipo_local,
			equipo_visitante,
			goles_equipo_local,
			goles_equipo_visitante,
			jornada,
			estadio
		FROM partidos
		WHERE id = $1
	`

	err := m.DB.QueryRow(
		query,
		id,
	).Scan(
		&partido.ID,
		&partido.EquipoLocal,
		&partido.EquipoVisitante,
		&partido.GolesEquipoLocal,
		&partido.GolesEquipoVisitante,
		&partido.Jornada,
		&partido.Estadio,
	)

	if err != nil {
		return nil, err
	}

	return partido, nil
}

// ACTUALIZAR
func (m *PartidoModel) Update(id string, partido *Partido) (*Partido, error) {

	query := `
		UPDATE partidos
		SET
			equipo_local = $1,
			equipo_visitante = $2,
			goles_equipo_local = $3,
			goles_equipo_visitante = $4,
			jornada = $5,
			estadio = $6
		WHERE id = $7
		RETURNING
			id,
			equipo_local,
			equipo_visitante,
			goles_equipo_local,
			goles_equipo_visitante,
			jornada,
			estadio
	`

	updatedPartido := &Partido{}

	err := m.DB.QueryRow(
		query,
		partido.EquipoLocal,
		partido.EquipoVisitante,
		partido.GolesEquipoLocal,
		partido.GolesEquipoVisitante,
		partido.Jornada,
		partido.Estadio,
		id,
	).Scan(
		&updatedPartido.ID,
		&updatedPartido.EquipoLocal,
		&updatedPartido.EquipoVisitante,
		&updatedPartido.GolesEquipoLocal,
		&updatedPartido.GolesEquipoVisitante,
		&updatedPartido.Jornada,
		&updatedPartido.Estadio,
	)

	if err != nil {
		return nil, err
	}

	return updatedPartido, nil
}

// BORRAR
func (m *PartidoModel) Delete(id string) error {

	query := `
		DELETE FROM partidos
		WHERE id = $1
	`

	result, err := m.DB.Exec(query, id)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
