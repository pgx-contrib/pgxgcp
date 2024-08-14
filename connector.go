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
	// Options to configure the GCP session
	Options []cloudsqlconn.Option
}

// BeforeConnect is called before a new connection is made. It is passed a copy of the underlying pgx.ConnConfig and
// will not impact any existing open connections.
func (x *Connector) BeforeConnect(ctx context.Context, conn *pgx.ConnConfig) error {
	// check if GOOGLE_APPLICATION_CREDENTIAL is set
	if os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
		return nil
	}

	// create a new dialer
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
