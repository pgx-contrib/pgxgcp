package pgxgcp

import (
	"context"
	"net"

	"cloud.google.com/go/cloudsqlconn"
	"github.com/jackc/pgx/v5"
)

// Connector connects to a Cloud SQL instance using the Cloud SQL Proxy.
type Connector struct {
	// Dialer is the underlying dialer used to connect to the Cloud SQL instance.
	Dialer *cloudsqlconn.Dialer
}

// Connect creates a new Connector using the provided options.
func Connect(ctx context.Context, options ...cloudsqlconn.Option) (*Connector, error) {
	// create a new dialer
	dialer, err := cloudsqlconn.NewDialer(ctx, options...)
	if err != nil {
		return nil, err
	}

	return &Connector{Dialer: dialer}, nil
}

// Close closes the connector and releases all resources held by the underlying dialer.
func (x *Connector) Close() error {
	return x.Dialer.Close()
}

// BeforeConnect is called before a new connection is made. It is passed a copy of the underlying pgx.ConnConfig and
// will not impact any existing open connections.
func (x *Connector) BeforeConnect(ctx context.Context, conn *pgx.ConnConfig) error {
	conn.DialFunc = func(ctx context.Context, _ string, _ string) (net.Conn, error) {
		// use the instance name from the host field
		return x.Dialer.Dial(ctx, conn.Host)
	}

	return nil
}
