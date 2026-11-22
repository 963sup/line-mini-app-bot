// Package objectstorage provides access to objects in a Google Cloud Storage bucket.
package objectstorage

import (
	"context"
	"errors"
	"io"
	"strings"

	"cloud.google.com/go/storage"
)

// Bucket uses a single Cloud Storage client for all operations on a bucket.
type Bucket struct {
	client *storage.Client
	bucket *storage.BucketHandle
}

// OpenBucket uses Application Default Credentials. The caller owns the client
// and must close it when the service shuts down.
func OpenBucket(ctx context.Context, name string) (*Bucket, error) {
	if name == "" || strings.ContainsAny(name, "/ \t\r\n") {
		return nil, errors.New("object storage requires a bucket name without a URL or path")
	}
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return &Bucket{client: client, bucket: client.Bucket(name)}, nil
}

func (b *Bucket) Close() error { return b.client.Close() }

// Upload streams the object and returns only after Cloud Storage confirms the
// write. An existing object with the same name is replaced.
func (b *Bucket) Upload(ctx context.Context, name, contentType string, src io.Reader) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	w := b.bucket.Object(name).NewWriter(ctx)
	w.ContentType = contentType
	if _, err := io.Copy(w, src); err != nil {
		cancel()
		_ = w.Close()
		return err
	}
	return w.Close()
}

// Open returns an object reader. The caller must close it.
func (b *Bucket) Open(ctx context.Context, name string) (io.ReadCloser, error) {
	return b.bucket.Object(name).NewReader(ctx)
}

func (b *Bucket) Stat(ctx context.Context, name string) (*storage.ObjectAttrs, error) {
	return b.bucket.Object(name).Attrs(ctx)
}

func (b *Bucket) Delete(ctx context.Context, name string) error {
	return b.bucket.Object(name).Delete(ctx)
}
