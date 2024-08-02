package pgxgcp

import (
	"context"
	"net"

	"cloud.google.com/go/cloudsqlconn"
	"github.com/jackc/pgx/v5"
)

// Connector connects to a Cloud SQL instance using the Cloud SQL Proxy.
type Connector struct {
	// Options to configure the GCP session
	Options []cloudsqlconn.Option
	// SkipAuth skips the authentication
	SkipAuth bool
}

// BeforeConnect is called before a new connection is made. It is passed a copy of the underlying pgx.ConnConfig and
// will not impact any existing open connections.
func (x *Connector) BeforeConnect(ctx context.Context, conn *pgx.ConnConfig) error {
	// skip any authentication
	if x.SkipAuth {
		return nil
	}

	dialer, err := cloudsqlconn.NewDialer(ctx, x.Options...)
	if err != nil {
		return err
	}

	conn.DialFunc = func(ctx context.Context, _ string, _ string) (net.Conn, error) {
		// we are considering the host name
		return dialer.Dial(ctx, conn.Host)
	}

	return nil
}
