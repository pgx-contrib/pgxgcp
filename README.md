# pgxgcp

[![CI](https://github.com/pgx-contrib/pgxgcp/actions/workflows/ci.yaml/badge.svg)](https://github.com/pgx-contrib/pgxgcp/actions/workflows/ci.yaml)
[![Release](https://img.shields.io/github/v/release/pgx-contrib/pgxgcp)](https://github.com/pgx-contrib/pgxgcp/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/pgx-contrib/pgxgcp.svg)](https://pkg.go.dev/github.com/pgx-contrib/pgxgcp)
[![License](https://img.shields.io/github/license/pgx-contrib/pgxgcp)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/pgx-contrib/pgxgcp)](go.mod)
[![pgx](https://img.shields.io/badge/pgx-v5-blue)](https://github.com/jackc/pgx)
[![GCP](https://img.shields.io/badge/GCP-Cloud_SQL-4285F4)](https://cloud.google.com/sql)

Google Cloud integration for [pgx v5](https://github.com/jackc/pgx), supporting Cloud SQL connections via the Cloud SQL Connector, plus query caching backends using Firestore, Datastore, and Cloud Storage.

## Features

- **Cloud SQL Connector** — connect to Cloud SQL instances via `pgx.ConnConfig.BeforeConnect` using the Cloud SQL Proxy dialer
- **FirestoreQueryCacher** — query result caching backed by Google Firestore (implements [pgxcache](https://github.com/pgx-contrib/pgxcache))
- **DatastoreQueryCacher** — query result caching backed by Google Datastore (implements [pgxcache](https://github.com/pgx-contrib/pgxcache))
- **StorageQueryCacher** — query result caching backed by Google Cloud Storage (implements [pgxcache](https://github.com/pgx-contrib/pgxcache))

## Installation

```bash
go get github.com/pgx-contrib/pgxgcp
```

## Usage

### Connector (Cloud SQL)

The `Connector` uses the Cloud SQL Proxy dialer to connect to Cloud SQL instances. It hooks into `pgx.ConnConfig.BeforeConnect` and replaces the dial function when `GOOGLE_APPLICATION_CREDENTIALS` is set.

```go
config, err := pgxpool.ParseConfig(os.Getenv("PGX_DATABASE_URL"))
if err != nil {
    panic(err)
}

ctx := context.TODO()
// Create a new pgxgcp.Connector
connector, err := pgxgcp.Connect(ctx)
if err != nil {
    panic(err)
}

// Set BeforeConnect hook to connector.BeforeConnect
config.BeforeConnect = connector.BeforeConnect

// Create a new pgxpool with the config
conn, err := pgxpool.NewWithConfig(ctx, config)
if err != nil {
    panic(err)
}

rows, err := conn.Query(ctx, "SELECT * from organization")
if err != nil {
    panic(err)
}
defer rows.Close()

type Organization struct {
    Name string `db:"name"`
}

for rows.Next() {
    organization, err := pgx.RowToStructByName[Organization](rows)
    if err != nil {
        panic(err)
    }
    fmt.Println(organization.Name)
}
```

### FirestoreQueryCacher

Cache query results in Google Firestore using [pgxcache](https://github.com/pgx-contrib/pgxcache):

```go
// Create a new client
client, err := firestore.NewClient(context.TODO(), os.Getenv("GOOGLE_PROJECT_ID"))
if err != nil {
    panic(err)
}

// Create a new cacher
cacher := &pgxgcp.FirestoreQueryCacher{
    Client:     client,
    Collection: "queries",
}

// create a new querier
querier := &pgxcache.Querier{
    Options: &pgxcache.QueryOptions{
        MaxLifetime: 30 * time.Second,
        MaxRows:     1,
    },
    Cacher:  cacher,
    Querier: conn,
}

rows, err := querier.Query(context.TODO(), "SELECT * from customer")
```

### DatastoreQueryCacher

Cache query results in Google Datastore using [pgxcache](https://github.com/pgx-contrib/pgxcache):

```go
// Create a new client
client, err := datastore.NewClient(context.TODO(), os.Getenv("GOOGLE_PROJECT_ID"))
if err != nil {
    panic(err)
}

// Create a new cacher
cacher := &pgxgcp.DatastoreQueryCacher{
    Client: client,
    Kind:   "queries",
}

// create a new querier
querier := &pgxcache.Querier{
    Options: &pgxcache.QueryOptions{
        MaxLifetime: 30 * time.Second,
        MaxRows:     1,
    },
    Cacher:  cacher,
    Querier: conn,
}

rows, err := querier.Query(context.TODO(), "SELECT * from customer")
```

### StorageQueryCacher

Cache query results in Google Cloud Storage using [pgxcache](https://github.com/pgx-contrib/pgxcache):

```go
// Create a new client
client, err := storage.NewClient(context.TODO())
if err != nil {
    panic(err)
}

// Create a new cacher
cacher := &pgxgcp.StorageQueryCacher{
    Client: client,
    Bucket: "queries",
}

// create a new querier
querier := &pgxcache.Querier{
    Options: &pgxcache.QueryOptions{
        MaxLifetime: 30 * time.Second,
        MaxRows:     1,
    },
    Cacher:  cacher,
    Querier: conn,
}

rows, err := querier.Query(context.TODO(), "SELECT * from customer")
```

## Development

### DevContainer

Open in VS Code with the Dev Containers extension. The environment provides Go,
PostgreSQL 18, and Nix automatically.

```
PGX_DATABASE_URL=postgres://vscode@postgres:5432/pgxgcp?sslmode=disable
```

### Nix

```bash
nix develop          # enter shell with Go
go tool ginkgo run -r
```

### Run tests

```bash
# Unit tests only (no database required)
go tool ginkgo run -r

# With integration tests
export PGX_DATABASE_URL="postgres://localhost/pgxgcp?sslmode=disable"
go tool ginkgo run -r
```

Integration tests require real GCP infrastructure and are guarded by environment variables — they are skipped automatically when the variables are not set:

| Variable | Used by |
|----------|---------|
| `PGXGCP_CLOUD_SQL_INSTANCE` | `Connector` (e.g. `project:region:instance`) |
| `GOOGLE_PROJECT_ID` | `FirestoreQueryCacher`, `DatastoreQueryCacher` |
| `PGXGCP_FIRESTORE_COLLECTION` | `FirestoreQueryCacher` |
| `PGXGCP_DATASTORE_KIND` | `DatastoreQueryCacher` |
| `PGXGCP_STORAGE_BUCKET` | `StorageQueryCacher` |

## License

[MIT](LICENSE)
