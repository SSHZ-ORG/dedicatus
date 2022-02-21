package scheduler

import (
	"context"
	"net/url"

	"github.com/SSHZ-ORG/dedicatus/paths"
	"github.com/SSHZ-ORG/dedicatus/scheduler/metadatamode"
	"google.golang.org/appengine/v2/taskqueue"
)

const updateFileMetadataQueue = "update-file-metadata"

func newUpdateFileMetadataTask(storageKey string, mode metadatamode.MetadataMode) *taskqueue.Task {
	return taskqueue.NewPOSTTask(paths.UpdateFileMetadata, url.Values{
		"id":   []string{storageKey},
		"mode": []string{mode.ToString()},
	})
}

func ScheduleUpdateFileMetadata(ctx context.Context, storageKeys []string, mode metadatamode.MetadataMode) error {
	var ts []*taskqueue.Task
	for _, id := range storageKeys {
		ts = append(ts, newUpdateFileMetadataTask(id, mode))
	}

	for _, batch := range batchize(ts) {
		_, err := taskqueue.AddMulti(ctx, batch, updateFileMetadataQueue)
		if err != nil {
			return err
		}
	}
	return nil
}
