package pgxgcp

import (
	"context"
	"time"

	"cloud.google.com/go/datastore"
	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
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
	// Client is the Firestore client.
	Client *firestore.Client
	// Collection is the name of the collection in Firestore.
	Collection string
}

// Get gets a cache item from Google Firestore. Returns pointer to the item, a boolean
// which represents whether key exists or not and an error.
func (r *FirestoreQueryCacher) Get(ctx context.Context, key *pgxcache.QueryKey) (*pgxcache.QueryResult, error) {
	// create a row
	row := &FirestoreQuery{
		ID: key.String(),
	}
	// get the item from the table
	document, err := r.Client.Collection(r.Collection).Doc(row.ID).Get(ctx)
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

	_, err = r.Client.Collection(r.Collection).Doc(row.ID).Set(ctx, row)
	return err
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
	// Client is the Datastore client.
	Client *datastore.Client
	// Kind is the name of the kind in Datastore.
	Kind string
}

// Get gets a cache item from Google Datastore. Returns pointer to the item, a boolean
// which represents whether key exists or not and an error.
func (r *DatastoreQueryCacher) Get(ctx context.Context, key *pgxcache.QueryKey) (*pgxcache.QueryResult, error) {
	// get the item from the collection
	row := &DatastoreQuery{
		ID: key.String(),
	}
	// create a new name
	name := datastore.NameKey(r.Kind, row.ID, nil)
	// get the item from the table
	err := r.Client.Get(ctx, name, row)
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
	name := datastore.NameKey(r.Kind, row.ID, nil)
	// set the item into the table
	_, err = r.Client.Put(ctx, name, row)
	// done!
	return err
}

var _ pgxcache.QueryCacher = &StorageQueryCacher{}

// StorageQueryCacher implements pgxcache.QueryCacher interface to use Google Cloud Storage.
type StorageQueryCacher struct {
	// Client to interact with S3
	Client *storage.Client
	// Bucket name in S3
	Bucket string
}

// Get implements pgxcache.QueryCacher.
func (r *StorageQueryCacher) Get(ctx context.Context, key *pgxcache.QueryKey) (*pgxcache.QueryResult, error) {
	// create a new entity
	entity := r.Client.Bucket(r.Bucket).Object(key.String())

	// check the expiration
	attr, err := entity.Attrs(ctx)
	switch err {
	case nil:
		if attr.CustomTime.Before(time.Now().UTC()) {
			return nil, nil
		}
	case storage.ErrObjectNotExist:
		return nil, nil
	default:
		return nil, err
	}

	// create a new writer
	reader, err := entity.NewReader(ctx)
	switch err {
	case nil:
		defer reader.Close()

		item := &pgxcache.QueryResult{}
		// unmarshal the result
		if err := msgpack.NewDecoder(reader).Decode(item); err != nil {
			return nil, err
		}
		// done!
		return item, nil
	case storage.ErrObjectNotExist:
		return nil, nil
	default:
		return nil, err
	}
}

// Set implements pgxcache.QueryCacher.
func (r *StorageQueryCacher) Set(ctx context.Context, key *pgxcache.QueryKey, item *pgxcache.QueryResult, ttl time.Duration) error {
	// create a new entity
	entity := r.Client.Bucket(r.Bucket).Object(key.String())
	// create a new writer
	writer := entity.NewWriter(ctx)
	// set the retention policy
	writer.ObjectAttrs.CustomTime = time.Now().UTC().Add(ttl)
	// close the writer
	defer writer.Close()

	// encode the item
	return msgpack.NewEncoder(writer).Encode(item)
}
