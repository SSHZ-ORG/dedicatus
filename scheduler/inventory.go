package scheduler

import (
	"net/url"

	"github.com/SSHZ-ORG/dedicatus/paths"
	"golang.org/x/net/context"
	"google.golang.org/appengine/taskqueue"
)

const updateFileMetadataQueue = "update-file-metadata"

func newUpdateFileMetadataTask(storageKey string) *taskqueue.Task {
	return taskqueue.NewPOSTTask(paths.UpdateFileMetadata, url.Values{
		"id": []string{storageKey},
	})
}

func ScheduleUpdateFileMetadata(ctx context.Context, storageKeys []string) error {
	var ts []*taskqueue.Task
	for _, id := range storageKeys {
		ts = append(ts, newUpdateFileMetadataTask(id))
	}

	for _, batch := range batchize(ts) {
		_, err := taskqueue.AddMulti(ctx, batch, updateFileMetadataQueue)
		if err != nil {
			return err
		}
	}
	return nil
}
