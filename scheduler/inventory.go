package scheduler

import (
	"net/url"

	"github.com/SSHZ-ORG/dedicatus/paths"
	"golang.org/x/net/context"
	"google.golang.org/appengine/taskqueue"
)

const updateFileMetadataQueue = "update-file-metadata"

func newUpdateFileMetadataTask(fileID string) *taskqueue.Task {
	return taskqueue.NewPOSTTask(paths.UpdateFileMetadata, url.Values{
		"id": []string{fileID},
	})
}

func ScheduleUpdateFileMetadata(ctx context.Context, fileIDs []string) error {
	var ts []*taskqueue.Task
	for _, id := range fileIDs {
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
