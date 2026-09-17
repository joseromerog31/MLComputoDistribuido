package models

import (
	"database/sql"
	"time"
)

type Partido struct {
	ID                 int64     `json:"id"`
	Referee            *string   `json:"referee"`
	Timezone           *string   `json:"timezone"`
	Date               time.Time `json:"date"`
	VenueID            *int64    `json:"venue_id"`
	VenueName          *string   `json:"venue_name"`
	VenueCity          *string   `json:"venue_city"`
	Season             int       `json:"season"`
	Round              *string   `json:"round"`
	HomeTeam           string    `json:"home_team"`
	AwayTeam           string    `json:"away_team"`
	HomeWin            *bool     `json:"home_win"`
	AwayWin            *bool     `json:"away_win"`
	HomeGoals          *int      `json:"home_goals"`
	AwayGoals          *int      `json:"away_goals"`
	HomeGoalsHalfTime  *int      `json:"home_goals_half_time"`
	AwayGoalsHalfTime  *int      `json:"away_goals_half_time"`
	HomeGoalsFulltime  *int      `json:"home_goals_fulltime"`
	AwayGoalsFulltime  *int      `json:"away_goals_fulltime"`
	HomeGoalsExtraTime *int      `json:"home_goals_extra_time"`
	AwayGoalsExtraTime *int      `json:"away_goals_extratime"`
	HomeGoalsPenalty   *int      `json:"home_goals_penalty"`
	AwayGoalsPenalty   *int      `json:"away_goals_penalty"`
}

type PartidoModel struct {
	DB *sql.DB
}

// Permite usar la misma función de Scan()
// tanto con QueryRow como con Rows.
type scanner interface {
	Scan(dest ...any) error
}

func scanPartido(s scanner, partido *Partido) error {
	return s.Scan(
		&partido.ID,
		&partido.Referee,
		&partido.Timezone,
		&partido.Date,
		&partido.VenueID,
		&partido.VenueName,
		&partido.VenueCity,
		&partido.Season,
		&partido.Round,
		&partido.HomeTeam,
		&partido.AwayTeam,
		&partido.HomeWin,
		&partido.AwayWin,
		&partido.HomeGoals,
		&partido.AwayGoals,
		&partido.HomeGoalsHalfTime,
		&partido.AwayGoalsHalfTime,
		&partido.HomeGoalsFulltime,
		&partido.AwayGoalsFulltime,
		&partido.HomeGoalsExtraTime,
		&partido.AwayGoalsExtraTime,
		&partido.HomeGoalsPenalty,
		&partido.AwayGoalsPenalty,
	)
}

// CREATE
func (m *PartidoModel) Create(partido *Partido) error {

	query := `
		INSERT INTO partidos (
			id,
			referee,
			timezone,
			date,
			venue_id,
			venue_name,
			venue_city,
			season,
			round,
			home_team,
			away_team,
			home_win,
			away_win,
			home_goals,
			away_goals,
			home_goals_half_time,
			away_goals_half_time,
			home_goals_fulltime,
			away_goals_fulltime,
			home_goals_extra_time,
			away_goals_extratime,
			home_goals_penalty,
			away_goals_penalty
		)
		VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20,
			$21, $22, $23
		)
	`

	_, err := m.DB.Exec(
		query,
		partido.ID,
		partido.Referee,
		partido.Timezone,
		partido.Date,
		partido.VenueID,
		partido.VenueName,
		partido.VenueCity,
		partido.Season,
		partido.Round,
		partido.HomeTeam,
		partido.AwayTeam,
		partido.HomeWin,
		partido.AwayWin,
		partido.HomeGoals,
		partido.AwayGoals,
		partido.HomeGoalsHalfTime,
		partido.AwayGoalsHalfTime,
		partido.HomeGoalsFulltime,
		partido.AwayGoalsFulltime,
		partido.HomeGoalsExtraTime,
		partido.AwayGoalsExtraTime,
		partido.HomeGoalsPenalty,
		partido.AwayGoalsPenalty,
	)

	return err
}

// READ ALL
func (m *PartidoModel) GetAll() ([]Partido, error) {

	query := `
		SELECT
			id,
			referee,
			timezone,
			date,
			venue_id,
			venue_name,
			venue_city,
			season,
			round,
			home_team,
			away_team,
			home_win,
			away_win,
			home_goals,
			away_goals,
			home_goals_half_time,
			away_goals_half_time,
			home_goals_fulltime,
			away_goals_fulltime,
			home_goals_extra_time,
			away_goals_extratime,
			home_goals_penalty,
			away_goals_penalty
		FROM partidos
		ORDER BY date, id
	`

	rows, err := m.DB.Query(query)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	partidos := []Partido{}

	for rows.Next() {

		var partido Partido

		err := scanPartido(
			rows,
			&partido,
		)

		if err != nil {
			return nil, err
		}

		partidos = append(
			partidos,
			partido,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return partidos, nil
}

// READ ONE
func (m *PartidoModel) GetByID(id string) (*Partido, error) {

	partido := &Partido{}

	query := `
		SELECT
			id,
			referee,
			timezone,
			date,
			venue_id,
			venue_name,
			venue_city,
			season,
			round,
			home_team,
			away_team,
			home_win,
			away_win,
			home_goals,
			away_goals,
			home_goals_half_time,
			away_goals_half_time,
			home_goals_fulltime,
			away_goals_fulltime,
			home_goals_extra_time,
			away_goals_extratime,
			home_goals_penalty,
			away_goals_penalty
		FROM partidos
		WHERE id = $1
	`

	err := scanPartido(
		m.DB.QueryRow(query, id),
		partido,
	)

	if err != nil {
		return nil, err
	}

	return partido, nil
}

// UPDATE
func (m *PartidoModel) Update(
	id string,
	partido *Partido,
) (*Partido, error) {

	query := `
		UPDATE partidos
		SET
			referee = $1,
			timezone = $2,
			date = $3,
			venue_id = $4,
			venue_name = $5,
			venue_city = $6,
			season = $7,
			round = $8,
			home_team = $9,
			away_team = $10,
			home_win = $11,
			away_win = $12,
			home_goals = $13,
			away_goals = $14,
			home_goals_half_time = $15,
			away_goals_half_time = $16,
			home_goals_fulltime = $17,
			away_goals_fulltime = $18,
			home_goals_extra_time = $19,
			away_goals_extratime = $20,
			home_goals_penalty = $21,
			away_goals_penalty = $22
		WHERE id = $23
		RETURNING
			id,
			referee,
			timezone,
			date,
			venue_id,
			venue_name,
			venue_city,
			season,
			round,
			home_team,
			away_team,
			home_win,
			away_win,
			home_goals,
			away_goals,
			home_goals_half_time,
			away_goals_half_time,
			home_goals_fulltime,
			away_goals_fulltime,
			home_goals_extra_time,
			away_goals_extratime,
			home_goals_penalty,
			away_goals_penalty
	`

	updatedPartido := &Partido{}

	err := scanPartido(
		m.DB.QueryRow(
			query,
			partido.Referee,
			partido.Timezone,
			partido.Date,
			partido.VenueID,
			partido.VenueName,
			partido.VenueCity,
			partido.Season,
			partido.Round,
			partido.HomeTeam,
			partido.AwayTeam,
			partido.HomeWin,
			partido.AwayWin,
			partido.HomeGoals,
			partido.AwayGoals,
			partido.HomeGoalsHalfTime,
			partido.AwayGoalsHalfTime,
			partido.HomeGoalsFulltime,
			partido.AwayGoalsFulltime,
			partido.HomeGoalsExtraTime,
			partido.AwayGoalsExtraTime,
			partido.HomeGoalsPenalty,
			partido.AwayGoalsPenalty,
			id,
		),
		updatedPartido,
	)

	if err != nil {
		return nil, err
	}

	return updatedPartido, nil
}

// DELETE
func (m *PartidoModel) Delete(id string) error {

	query := `
		DELETE FROM partidos
		WHERE id = $1
	`

	result, err := m.DB.Exec(
		query,
		id,
	)

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
