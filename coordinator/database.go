package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func connectDatabase() (*sql.DB, error) {

	databaseURL :=
		os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		return nil, fmt.Errorf(
			"DATABASE_URL no está definida",
		)
	}

	db, err := sql.Open(
		"pgx",
		databaseURL,
	)

	if err != nil {
		return nil, err
	}

	// Timeout para comprobar conexión
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

	// Pool sencillo para este proyecto.
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	return db, nil
}
