package pgxgcp_test

import (
	"context"
	"fmt"
	"os"
	"time"

	"cloud.google.com/go/datastore"
	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pgx-contrib/pgxcache"
	"github.com/pgx-contrib/pgxgcp"
)

var _ = Describe("FirestoreQueryCacher", func() {
	// -------------------------------------------------------------------------
	Describe("Integration", Ordered, func() {
		var (
			client *firestore.Client
			cacher *pgxgcp.FirestoreQueryCacher
			ctx    context.Context
			key    *pgxcache.QueryKey
		)

		BeforeAll(func() {
			project := os.Getenv("GOOGLE_PROJECT_ID")
			collection := os.Getenv("PGXGCP_FIRESTORE_COLLECTION")
			if project == "" || collection == "" {
				Skip("GOOGLE_PROJECT_ID and PGXGCP_FIRESTORE_COLLECTION must be set")
			}

			var err error
			client, err = firestore.NewClient(context.Background(), project)
			Expect(err).NotTo(HaveOccurred())

			cacher = &pgxgcp.FirestoreQueryCacher{
				Client:     client,
				Collection: collection,
			}
			ctx = context.Background()
			key = &pgxcache.QueryKey{SQL: fmt.Sprintf("SELECT '%d'", time.Now().UnixNano())}
		})

		AfterAll(func() {
			if client != nil {
				client.Close()
			}
		})

		It("returns nil for a missing key", func() {
			item, err := cacher.Get(ctx, key)
			Expect(err).NotTo(HaveOccurred())
			Expect(item).To(BeNil())
		})

		It("round-trips an item through Set and Get", func() {
			item := &pgxcache.QueryItem{CommandTag: "SELECT"}
			Expect(cacher.Set(ctx, key, item, time.Minute)).To(Succeed())

			got, err := cacher.Get(ctx, key)
			Expect(err).NotTo(HaveOccurred())
			Expect(got).NotTo(BeNil())
			Expect(got.CommandTag).To(Equal("SELECT"))
		})

		It("returns nil for an expired item", func() {
			expiredKey := &pgxcache.QueryKey{SQL: fmt.Sprintf("SELECT 'expired-%d'", time.Now().UnixNano())}
			item := &pgxcache.QueryItem{CommandTag: "SELECT"}
			// Negative TTL places ExpireAt in the past, so Get treats it as a cache miss.
			Expect(cacher.Set(ctx, expiredKey, item, -time.Second)).To(Succeed())

			got, err := cacher.Get(ctx, expiredKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(BeNil())
		})

		It("Reset returns nil", func() {
			Expect(cacher.Reset(ctx)).To(Succeed())
		})
	})
})

var _ = Describe("DatastoreQueryCacher", func() {
	// -------------------------------------------------------------------------
	Describe("Integration", Ordered, func() {
		var (
			client *datastore.Client
			cacher *pgxgcp.DatastoreQueryCacher
			ctx    context.Context
			key    *pgxcache.QueryKey
		)

		BeforeAll(func() {
			project := os.Getenv("GOOGLE_PROJECT_ID")
			kind := os.Getenv("PGXGCP_DATASTORE_KIND")
			if project == "" || kind == "" {
				Skip("GOOGLE_PROJECT_ID and PGXGCP_DATASTORE_KIND must be set")
			}

			var err error
			client, err = datastore.NewClient(context.Background(), project)
			Expect(err).NotTo(HaveOccurred())

			cacher = &pgxgcp.DatastoreQueryCacher{
				Client: client,
				Kind:   kind,
			}
			ctx = context.Background()
			key = &pgxcache.QueryKey{SQL: fmt.Sprintf("SELECT '%d'", time.Now().UnixNano())}
		})

		AfterAll(func() {
			if client != nil {
				client.Close()
			}
		})

		It("returns nil for a missing key", func() {
			item, err := cacher.Get(ctx, key)
			Expect(err).NotTo(HaveOccurred())
			Expect(item).To(BeNil())
		})

		It("round-trips an item through Set and Get", func() {
			item := &pgxcache.QueryItem{CommandTag: "SELECT"}
			Expect(cacher.Set(ctx, key, item, time.Minute)).To(Succeed())

			got, err := cacher.Get(ctx, key)
			Expect(err).NotTo(HaveOccurred())
			Expect(got).NotTo(BeNil())
			Expect(got.CommandTag).To(Equal("SELECT"))
		})

		It("returns nil for an expired item", func() {
			expiredKey := &pgxcache.QueryKey{SQL: fmt.Sprintf("SELECT 'expired-%d'", time.Now().UnixNano())}
			item := &pgxcache.QueryItem{CommandTag: "SELECT"}
			// Negative TTL places ExpireAt in the past, so Get treats it as a cache miss.
			Expect(cacher.Set(ctx, expiredKey, item, -time.Second)).To(Succeed())

			got, err := cacher.Get(ctx, expiredKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(BeNil())
		})

		It("Reset returns nil", func() {
			Expect(cacher.Reset(ctx)).To(Succeed())
		})
	})
})

var _ = Describe("StorageQueryCacher", func() {
	// -------------------------------------------------------------------------
	Describe("Integration", Ordered, func() {
		var (
			client *storage.Client
			cacher *pgxgcp.StorageQueryCacher
			ctx    context.Context
			key    *pgxcache.QueryKey
		)

		BeforeAll(func() {
			bucket := os.Getenv("PGXGCP_STORAGE_BUCKET")
			if bucket == "" {
				Skip("PGXGCP_STORAGE_BUCKET must be set")
			}

			var err error
			client, err = storage.NewClient(context.Background())
			Expect(err).NotTo(HaveOccurred())

			cacher = &pgxgcp.StorageQueryCacher{
				Client: client,
				Bucket: bucket,
			}
			ctx = context.Background()
			key = &pgxcache.QueryKey{SQL: fmt.Sprintf("SELECT '%d'", time.Now().UnixNano())}
		})

		AfterAll(func() {
			if client != nil {
				client.Close()
			}
		})

		It("returns nil for a missing key", func() {
			item, err := cacher.Get(ctx, key)
			Expect(err).NotTo(HaveOccurred())
			Expect(item).To(BeNil())
		})

		It("round-trips an item through Set and Get", func() {
			item := &pgxcache.QueryItem{CommandTag: "SELECT"}
			Expect(cacher.Set(ctx, key, item, time.Minute)).To(Succeed())

			got, err := cacher.Get(ctx, key)
			Expect(err).NotTo(HaveOccurred())
			Expect(got).NotTo(BeNil())
			Expect(got.CommandTag).To(Equal("SELECT"))
		})

		It("returns nil for an expired item", func() {
			expiredKey := &pgxcache.QueryKey{SQL: fmt.Sprintf("SELECT 'expired-%d'", time.Now().UnixNano())}
			item := &pgxcache.QueryItem{CommandTag: "SELECT"}
			// Negative TTL places CustomTime in the past, so Get treats it as a cache miss.
			Expect(cacher.Set(ctx, expiredKey, item, -time.Second)).To(Succeed())

			got, err := cacher.Get(ctx, expiredKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(BeNil())
		})

		It("Reset returns nil", func() {
			Expect(cacher.Reset(ctx)).To(Succeed())
		})
	})
})
