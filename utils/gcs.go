package utils

import (
	"context"

	"cloud.google.com/go/storage"
	"google.golang.org/appengine/v2"
)

const gcsFilePrefix = "TgFileV2/"

func StoreFileToGCS(ctx context.Context, fileUniqueID string, bytes []byte) error {
	sc, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}

	w := sc.Bucket(appengine.DefaultVersionHostname(ctx)).Object(gcsFilePrefix + fileUniqueID + ".mp4").NewWriter(ctx)
	if _, err = w.Write(bytes); err != nil {
		return err
	}
	return w.Close()
}
