package pgxgcp_test

import (
	"context"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgx-contrib/pgxgcp"
)

var count int

func ExampleConnector() {
	config, err := pgxpool.ParseConfig(os.Getenv("PGX_DATABASE_URL"))
	if err != nil {
		panic(err)
	}

	// Set max connection lifetime to 5 minutes in postgres connection pool configuration.
	// Note: this will refresh connections before the 15 min expiration on the IAM AWS auth token,
	// while leveraging the BeforeConnect hook to recreate the token in time dynamically.
	config.MaxConnLifetime = 5 * time.Minute

	// Create a new pgxaws.Connector
	connector := &pgxgcp.Connector{}
	// Set BeforeConnect hook to pgxaws.BeforeConnect
	config.BeforeConnect = connector.BeforeConnect

	conn, err := pgxpool.NewWithConfig(context.TODO(), config)
	if err != nil {
		panic(err)
	}

	row := conn.QueryRow(context.TODO(), "SELECT 1")
	if err := row.Scan(&count); err != nil {
		panic(err)
	}
}
