package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Crea y valida la conexión con Postgres
func connectDatabase() (*sql.DB, error) {

	// Variable de entorno para el string de la conexión
	databaseURL :=
		os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		return nil, fmt.Errorf(
			"DATABASE_URL no está definida",
		)
	}

	// pgx se usa como driver de Postgres
	db, err := sql.Open(
		"pgx",
		databaseURL,
	)

	if err != nil {
		return nil, err
	}

	// Timeout para comprobar conexión y que no se acepten requests sin acceso
	ctx, cancel :=
		context.WithTimeout(
			context.Background(),
			5*time.Second,
		)

	defer cancel()

	err = db.PingContext(ctx)

	if err != nil {
		db.Close()

		return nil, fmt.Errorf(
			"no se pudo conectar a PostgreSQL: %w",
			err,
		)
	}

	// No se ocupan más conexiones
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	return db, nil
}
