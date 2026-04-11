package pgxgcp_test

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pgx-contrib/pgxgcp"
)

var _ = Describe("Connector", func() {
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
	})

	// -------------------------------------------------------------------------
	Describe("BeforeConnect", func() {
		It("sets DialFunc on the conn config", func() {
			connector := &pgxgcp.Connector{}

			conn := &pgx.ConnConfig{}
			conn.Host = "project:region:instance"

			Expect(connector.BeforeConnect(ctx, conn)).To(Succeed())
			Expect(conn.DialFunc).NotTo(BeNil())
		})

		It("always returns nil", func() {
			connector := &pgxgcp.Connector{}

			conn := &pgx.ConnConfig{}
			Expect(connector.BeforeConnect(ctx, conn)).To(Succeed())
		})
	})

	// -------------------------------------------------------------------------
	Describe("Integration", Ordered, func() {
		var connector *pgxgcp.Connector

		BeforeAll(func() {
			if os.Getenv("PGXGCP_CLOUD_SQL_INSTANCE") == "" {
				Skip("PGXGCP_CLOUD_SQL_INSTANCE not set")
			}

			var err error
			connector, err = pgxgcp.Connect(ctx)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterAll(func() {
			if connector != nil {
				Expect(connector.Close()).To(Succeed())
			}
		})

		It("Connect returns a Connector with a non-nil Dialer", func() {
			Expect(connector.Dialer).NotTo(BeNil())
		})

		It("BeforeConnect sets DialFunc for the Cloud SQL instance", func() {
			conn := &pgx.ConnConfig{}
			conn.Host = os.Getenv("PGXGCP_CLOUD_SQL_INSTANCE")

			Expect(connector.BeforeConnect(ctx, conn)).To(Succeed())
			Expect(conn.DialFunc).NotTo(BeNil())
		})

		It("Close succeeds without error", func() {
			Expect(connector.Close()).To(Succeed())
			connector = nil
		})
	})
})
