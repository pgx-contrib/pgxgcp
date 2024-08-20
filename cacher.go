package pgxgcp

import (
	"context"
	"time"

	"cloud.google.com/go/datastore"
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
	// create a row
	row := &FirestoreQuery{
		ID: key.String(),
	}
	// get the item from the table
	document, err := r.client.Collection(r.collection).Doc(row.ID).Get(ctx)
	switch status.Code(err) {
	case codes.OK:
		// get the record
		if err := document.DataTo(row); err != nil {
			return nil, err
		}

		item := &pgxcache.QueryResult{}
		// unmarshal the result
		if err := msgpack.Unmarshal(row.Data, item); err != nil {
			return nil, err
		}
		return item, nil
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

	_, err = r.client.Collection(r.collection).Doc(row.ID).Set(ctx, row)
	return err
}

// Close closes the FirestoreQueryCacher client.
func (r *FirestoreQueryCacher) Close() error {
	return r.client.Close()
}

// DatastoreQuery represents a record in the dynamodb table.
type DatastoreQuery struct {
	ID       string    `datastore:"query_id,hash"`
	Data     []byte    `datastore:"query_data"`
	ExpireAt time.Time `datastore:"query_expire_at"`
}

var _ pgxcache.QueryCacher = &DatastoreQueryCacher{}

// DatastoreQueryCacher implements pgxcache.QueryCacher interface to use Google Datastore.
type DatastoreQueryCacher struct {
	client *datastore.Client
	kind   string
}

// NewDatastoreQueryCacher creates a new instance of DatastoreQueryCacher backend using Google Datastore client.
// All rows created in Google Datastore by pgxcache will have stored with table.
func NewDatastoreQueryCacher(client *datastore.Client, kind string) *DatastoreQueryCacher {
	return &DatastoreQueryCacher{
		client: client,
		kind:   kind,
	}
}

// Get gets a cache item from Google Datastore. Returns pointer to the item, a boolean
// which represents whether key exists or not and an error.
func (r *DatastoreQueryCacher) Get(ctx context.Context, key *pgxcache.QueryKey) (*pgxcache.QueryResult, error) {
	// get the item from the collection
	row := &DatastoreQuery{
		ID: key.String(),
	}
	// create a new name
	name := datastore.NameKey(r.kind, row.ID, nil)
	// get the item from the table
	err := r.client.Get(ctx, name, row)
	switch err {
	case nil:
		item := &pgxcache.QueryResult{}
		// unmarshal the result
		if err := msgpack.Unmarshal(row.Data, item); err != nil {
			return nil, err
		}
		return item, nil
	case datastore.ErrNoSuchEntity:
		return nil, nil
	default:
		return nil, err
	}
}

// Set sets the given item into Google Datastore with provided TTL duration.
func (r *DatastoreQueryCacher) Set(ctx context.Context, key *pgxcache.QueryKey, item *pgxcache.QueryResult, ttl time.Duration) error {
	// marshal the item
	data, err := msgpack.Marshal(item)
	if err != nil {
		return err
	}

	// prepare the record
	row := &DatastoreQuery{
		ID:       key.String(),
		Data:     data,
		ExpireAt: time.Now().UTC().Add(ttl),
	}
	// create a new name
	name := datastore.NameKey(r.kind, row.ID, nil)
	// set the item into the table
	_, err = r.client.Put(ctx, name, row)
	// done!
	return err
}

// Close closes the FirestoreQueryCacher client.
func (r *DatastoreQueryCacher) Close() error {
	return r.client.Close()
}