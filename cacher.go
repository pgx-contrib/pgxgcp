package pgxgcp

import (
	"context"
	"io"
	"time"

	"cloud.google.com/go/datastore"
	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	"github.com/pgx-contrib/pgxcache"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// FirestoreQuery represents a record in a Firestore collection.
type FirestoreQuery struct {
	ID       string    `firestore:"-"`
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
func (r *FirestoreQueryCacher) Get(ctx context.Context, key *pgxcache.QueryKey) (*pgxcache.QueryItem, error) {
	// create a row
	row := &FirestoreQuery{
		ID: key.String(),
	}
	// get the item from the collection
	document, err := r.Client.Collection(r.Collection).Doc(row.ID).Get(ctx)
	switch status.Code(err) {
	case codes.OK:
		// get the record
		if err := document.DataTo(row); err != nil {
			return nil, err
		}

		// check if the item has expired
		if row.ExpireAt.Before(time.Now().UTC()) {
			return nil, nil
		}

		item := &pgxcache.QueryItem{}
		// unmarshal the result
		if err := item.UnmarshalText(row.Data); err != nil {
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
func (r *FirestoreQueryCacher) Set(ctx context.Context, key *pgxcache.QueryKey, item *pgxcache.QueryItem, ttl time.Duration) error {
	// marshal the item
	data, err := item.MarshalText()
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

// Reset implements pgxcache.QueryCacher.
func (r *FirestoreQueryCacher) Reset(context.Context) error {
	// TODO: implement this method
	return nil
}

// DatastoreQuery represents a record in a Datastore kind.
type DatastoreQuery struct {
	ID       string    `datastore:"-"`
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
func (r *DatastoreQueryCacher) Get(ctx context.Context, key *pgxcache.QueryKey) (*pgxcache.QueryItem, error) {
	// get the item from the kind
	row := &DatastoreQuery{
		ID: key.String(),
	}
	// create a new name key
	name := datastore.NameKey(r.Kind, row.ID, nil)
	// get the item from Datastore
	err := r.Client.Get(ctx, name, row)
	switch err {
	case nil:
		// check if the item has expired
		if row.ExpireAt.Before(time.Now().UTC()) {
			return nil, nil
		}

		item := &pgxcache.QueryItem{}
		// unmarshal the result
		if err := item.UnmarshalText(row.Data); err != nil {
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
func (r *DatastoreQueryCacher) Set(ctx context.Context, key *pgxcache.QueryKey, item *pgxcache.QueryItem, ttl time.Duration) error {
	// marshal the item
	data, err := item.MarshalText()
	if err != nil {
		return err
	}

	// prepare the record
	row := &DatastoreQuery{
		ID:       key.String(),
		Data:     data,
		ExpireAt: time.Now().UTC().Add(ttl),
	}
	// create a new name key
	name := datastore.NameKey(r.Kind, row.ID, nil)
	// set the item into Datastore
	_, err = r.Client.Put(ctx, name, row)
	return err
}

// Reset implements pgxcache.QueryCacher.
func (r *DatastoreQueryCacher) Reset(context.Context) error {
	// TODO: implement this method
	return nil
}

var _ pgxcache.QueryCacher = &StorageQueryCacher{}

// StorageQueryCacher implements pgxcache.QueryCacher interface to use Google Cloud Storage.
type StorageQueryCacher struct {
	// Client is the Cloud Storage client.
	Client *storage.Client
	// Bucket is the name of the Cloud Storage bucket.
	Bucket string
}

// Get implements pgxcache.QueryCacher.
func (r *StorageQueryCacher) Get(ctx context.Context, key *pgxcache.QueryKey) (*pgxcache.QueryItem, error) {
	// create a new entity
	entity := r.Client.Bucket(r.Bucket).Object(key.String())

	// check the expiration via CustomTime
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

	// read the object data
	reader, err := entity.NewReader(ctx)
	switch err {
	case nil:
		defer reader.Close()
		// read the data
		data, err := io.ReadAll(reader)
		if err != nil {
			return nil, err
		}

		item := &pgxcache.QueryItem{}
		// unmarshal the result
		if err := item.UnmarshalText(data); err != nil {
			return nil, err
		}
		return item, nil
	case storage.ErrObjectNotExist:
		return nil, nil
	default:
		return nil, err
	}
}

// Set implements pgxcache.QueryCacher.
func (r *StorageQueryCacher) Set(ctx context.Context, key *pgxcache.QueryKey, item *pgxcache.QueryItem, ttl time.Duration) error {
	// create a cancellable context so the upload is aborted on any error path
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// create a new entity
	entity := r.Client.Bucket(r.Bucket).Object(key.String())
	// create a new writer
	writer := entity.NewWriter(ctx)
	// set the expiry via CustomTime; the upload is only committed on Close
	writer.CustomTime = time.Now().UTC().Add(ttl)

	data, err := item.MarshalText()
	if err != nil {
		return err
	}

	if _, err = writer.Write(data); err != nil {
		return err
	}

	// Close finalises and commits the upload; its error must not be discarded
	return writer.Close()
}

// Reset implements pgxcache.QueryCacher.
func (r *StorageQueryCacher) Reset(context.Context) error {
	// TODO: implement this method
	return nil
}
