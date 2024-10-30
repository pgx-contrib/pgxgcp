package pgxgcp

import (
	"context"
	"net"
	"os"

	"cloud.google.com/go/cloudsqlconn"
	"github.com/jackc/pgx/v5"
)

// Connector connects to a Cloud SQL instance using the Cloud SQL Proxy.
type Connector struct {
	// dialer is the underlying dialer used to connect to the Cloud SQL instance.
	dialer *cloudsqlconn.Dialer
}

// Connect creates a new Connector using the provided options.
func Connect(ctx context.Context, options ...cloudsqlconn.Option) (*Connector, error) {
	// create a new dialer
	dialer, err := cloudsqlconn.NewDialer(ctx, options...)
	if err != nil {
		return nil, err
	}

	return &Connector{dialer: dialer}, nil
}

// BeforeConnect is called before a new connection is made. It is passed a copy of the underlying pgx.ConnConfig and
// will not impact any existing open connections.
func (x *Connector) BeforeConnect(ctx context.Context, conn *pgx.ConnConfig) error {
	// check if GOOGLE_APPLICATION_CREDENTIAL is set
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
		return nil
	}

	conn.DialFunc = func(ctx context.Context, _ string, _ string) (net.Conn, error) {
		// we are considering the host name
		return x.dialer.Dial(ctx, conn.Host)
	}

	return nil
}
