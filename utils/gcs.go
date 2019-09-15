package utils

import (
	"cloud.google.com/go/storage"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
)

const gcsFilePrefix = "TgFile/"

func StoreFileToGCS(ctx context.Context, fileID string, bytes []byte) error {
	sc, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}

	w := sc.Bucket(appengine.DefaultVersionHostname(ctx)).Object(gcsFilePrefix + fileID + ".mp4").NewWriter(ctx)
	if _, err = w.Write(bytes); err != nil {
		return err
	}
	return w.Close()
}
