package pgxgcp

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/pgx-contrib/pgxcache"
	"github.com/vmihailenco/msgpack/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// FirestoreQuery represents a record in the dynamodb table.
type FirestoreQuery struct {
	ID       string    `firestore:"query_id,hash"`
	Data     []byte    `firestore:"query_data"`
	ExpireAt time.Time `firestore:"query_expire_at"`
}

var _ pgxcache.QueryCacher = &FirestoreQueryCacher{}

// FirestoreQueryCacher implements pgxcache.QueryCacher interface to use Google Firestore.
type FirestoreQueryCacher struct {
	client     *firestore.Client
	collection string
}

// NewFirestoreQueryCacher creates a new instance of FirestoreQueryCacher backend using Google Firestore client.
// All rows created in Google Firestore by pgxcache will have stored with table.
func NewFirestoreQueryCacher(client *firestore.Client, collection string) *FirestoreQueryCacher {
	return &FirestoreQueryCacher{
		client:     client,
		collection: collection,
	}
}

// Get gets a cache item from Google Firestore. Returns pointer to the item, a boolean
// which represents whether key exists or not and an error.
func (r *FirestoreQueryCacher) Get(ctx context.Context, key *pgxcache.QueryKey) (*pgxcache.QueryResult, error) {
	// get the item from the table
	document, err := r.client.Collection(r.collection).Doc(key.String()).Get(ctx)
	switch status.Code(err) {
	case codes.OK:
		row := &FirestoreQuery{}
		// get the record
		if err := document.DataTo(row); err != nil {
			return nil, err
		}

		var item pgxcache.QueryResult
		// unmarshal the result
		if err := msgpack.Unmarshal(row.Data, &item); err != nil {
			return nil, err
		}
		return &item, nil
	case codes.NotFound:
		return nil, nil
	default:
		return nil, err
	}
}

// Set sets the given item into Google Firestore with provided TTL duration.
func (r *FirestoreQueryCacher) Set(ctx context.Context, key *pgxcache.QueryKey, item *pgxcache.QueryResult, ttl time.Duration) error {
	// marshal the item
	data, err := msgpack.Marshal(item)
	if err != nil {
		return err
	}

	// prepare the record
	row := &FirestoreQuery{
		ID:       key.String(),
		Data:     data,
		ExpireAt: time.Now().UTC().Add(ttl),
	}

	_, err = r.client.Collection(r.collection).Doc(key.String()).Set(ctx, row)
	return err
}
