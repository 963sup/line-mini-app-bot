package automatic

import "context"

// SourceStore addresses immutable workbook bytes by tenant-scoped owner and SHA-256.
type SourceStore interface {
	Put(ctx context.Context, repository, actor, digest string, data []byte) error
	Read(ctx context.Context, repository, actor, digest string) ([]byte, error)
}
