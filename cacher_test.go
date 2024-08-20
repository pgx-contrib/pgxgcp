package pgxgcp_test

import (
	"context"
	"os"
	"time"

	"cloud.google.com/go/datastore"
	"cloud.google.com/go/firestore"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgx-contrib/pgxcache"
	"github.com/pgx-contrib/pgxgcp"
)

func ExampleFirestoreQueryCacher() {
	config, err := pgxpool.ParseConfig(os.Getenv("PGX_DATABASE_URL"))
	if err != nil {
		panic(err)
	}

	conn, err := pgxpool.NewWithConfig(context.TODO(), config)
	if err != nil {
		panic(err)
	}

	// Create a new client
	client, err := firestore.NewClient(context.TODO(), os.Getenv("GOOGLE_PROJECT_ID"))
	if err != nil {
		panic(err)
	}

	// Create a new cacher
	cacher := pgxgcp.NewFirestoreQueryCacher(client, "queries")

	// create a new querier
	querier := &pgxcache.Querier{
		// set the default query options, which can be overridden by the query
		// -- @cache-max-rows 100
		// -- @cache-ttl 30s
		Options: &pgxcache.QueryOptions{
			MaxLifetime: 30 * time.Second,
			MaxRows:     1,
		},
		Cacher:  cacher,
		Querier: conn,
	}

	row := querier.QueryRow(context.TODO(), "SELECT 1")
	// scan the row
	if err := row.Scan(&count); err != nil {
		panic(err)
	}
}

func ExampleDatastoreQueryCacher() {
	config, err := pgxpool.ParseConfig(os.Getenv("PGX_DATABASE_URL"))
	if err != nil {
		panic(err)
	}

	conn, err := pgxpool.NewWithConfig(context.TODO(), config)
	if err != nil {
		panic(err)
	}

	// Create a new client
	client, err := datastore.NewClient(context.TODO(), os.Getenv("GOOGLE_PROJECT_ID"))
	if err != nil {
		panic(err)
	}

	// Create a new cacher
	cacher := pgxgcp.NewDatastoreQueryCacher(client, "queries")

	// create a new querier
	querier := &pgxcache.Querier{
		// set the default query options, which can be overridden by the query
		// -- @cache-max-rows 100
		// -- @cache-ttl 30s
		Options: &pgxcache.QueryOptions{
			MaxLifetime: 30 * time.Second,
			MaxRows:     1,
		},
		Cacher:  cacher,
		Querier: conn,
	}

	row := querier.QueryRow(context.TODO(), "SELECT 1")
	// scan the row
	if err := row.Scan(&count); err != nil {
		panic(err)
	}
}
