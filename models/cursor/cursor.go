package cursor

import (
	"strconv"
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/memcache"
)

const (
	memcachePrefix = "D545:QO1:"
)

func tryGetCursor(ctx context.Context, queryID string) *datastore.Cursor {
	i, err := memcache.Get(ctx, memcachePrefix+queryID)
	if err != nil {
		return nil
	}
	c, err := datastore.DecodeCursor(string(i.Value))
	if err != nil {
		return nil
	}
	return &c
}

func Offset(ctx context.Context, q *datastore.Query, lastCursor string) (*datastore.Query, int) {
	if lastCursor == "" {
		return q, 0
	}

	split := strings.SplitN(lastCursor, ":", 2)
	thisOffset := split[0]
	lastQueryID := split[1]

	offset, _ := strconv.Atoi(thisOffset)

	cursor := tryGetCursor(ctx, lastQueryID)
	if cursor != nil {
		return q.Start(*cursor), offset
	}

	return q.Offset(offset), offset
}

func Store(ctx context.Context, cursor datastore.Cursor, queryID string, nextOffset int) string {
	_ = memcache.Add(ctx, &memcache.Item{
		Key:   memcachePrefix + queryID,
		Value: []byte(cursor.String()),
	})

	return strconv.Itoa(nextOffset) + ":" + queryID
}
