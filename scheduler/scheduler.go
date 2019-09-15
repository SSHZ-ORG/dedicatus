package scheduler

import "google.golang.org/appengine/taskqueue"

const maxTasksPerAddMulti = 100

func batchize(ts []*taskqueue.Task) [][]*taskqueue.Task {
	var batches [][]*taskqueue.Task
	for maxTasksPerAddMulti < len(ts) {
		ts, batches = ts[maxTasksPerAddMulti:], append(batches, ts[0:maxTasksPerAddMulti:maxTasksPerAddMulti])
	}
	if len(ts) > 0 {
		batches = append(batches, ts)
	}
	return batches
}
