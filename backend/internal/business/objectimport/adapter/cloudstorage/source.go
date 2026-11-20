package cloudstorage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"

	"github.com/963sup/line-mini-app-bot/internal/platform/firestorekey"
	"google.golang.org/api/googleapi"
	storage "google.golang.org/api/storage/v1"
)

type Sources struct {
	Client          *storage.Service
	Bucket, Channel string
}

func (s Sources) key(repository, actor, digest string) (string, error) {
	hash, err := hex.DecodeString(digest)
	if err != nil || len(hash) != sha256.Size || digest != hex.EncodeToString(hash) || actor == "" || repository == "" || s.Channel == "" {
		return "", errors.New("invalid source reference")
	}
	return "object-import/" + firestorekey.Digest([]byte(s.Channel)) + "/" + firestorekey.Digest([]byte(repository)) + "/" + firestorekey.Digest([]byte(actor)) + "/" + digest, nil
}

func (s Sources) Put(ctx context.Context, repo, actor, digest string, data []byte) error {
	key, err := s.key(repo, actor, digest)
	if err != nil {
		return err
	}
	h := sha256.Sum256(data)
	if hex.EncodeToString(h[:]) != digest || len(data) > 300<<10 {
		return errors.New("invalid workbook bytes")
	}
	_, err = s.Client.Objects.Insert(s.Bucket, &storage.Object{Name: key}).IfGenerationMatch(0).Media(bytes.NewReader(data), googleapi.ContentType("application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")).Context(ctx).Do()
	var conflict *googleapi.Error
	if errors.As(err, &conflict) && conflict.Code == 412 {
		return nil
	}
	return err
}

func (s Sources) Read(ctx context.Context, repo, actor, digest string) ([]byte, error) {
	key, err := s.key(repo, actor, digest)
	if err != nil {
		return nil, err
	}
	response, err := s.Client.Objects.Get(s.Bucket, key).Context(ctx).Download()
	if err != nil {
		return nil, err
	}
	defer func() { _ = response.Body.Close() }()
	data, err := io.ReadAll(io.LimitReader(response.Body, (300<<10)+1))
	if err != nil {
		return nil, err
	}
	h := sha256.Sum256(data)
	if len(data) > 300<<10 || hex.EncodeToString(h[:]) != digest {
		return nil, errors.New("source content mismatch")
	}
	return data, nil
}
