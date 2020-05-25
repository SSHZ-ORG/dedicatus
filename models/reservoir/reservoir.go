package reservoir

import (
	"bytes"
	"math/rand"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/memcache"
)

const memcacheKey = "D545:IR2"

var separator = []byte(":")

// We don't use Gob here because Gob wastes space for things like field names in Go struct.

func TryRotateReservoir(ctx context.Context, limit int) bool {
	i, err := memcache.Get(ctx, memcacheKey)
	if err != nil {
		return false
	}

	ks := bytes.SplitN(i.Value, separator, 2*limit+1)

	if len(ks) < 2*limit {
		return false
	}

	ks = ks[limit:]
	err = memcache.Set(ctx, &memcache.Item{
		Key:   memcacheKey,
		Value: bytes.Join(ks, separator),
	})
	return err == nil
}

func RefillReservoir(ctx context.Context, keys []*datastore.Key) error {
	rand.Shuffle(len(keys), func(i, j int) {
		keys[i], keys[j] = keys[j], keys[i]
	})

	var bs [][]byte
	for _, k := range keys {
		bs = append(bs, []byte(k.Encode()))
	}

	return memcache.Set(ctx, &memcache.Item{
		Key:   memcacheKey,
		Value: bytes.Join(bs, separator),
	})
}

func ReadReservoir(ctx context.Context, limit int) ([]*datastore.Key, error) {
	i, err := memcache.Get(ctx, memcacheKey)
	if err == memcache.ErrCacheMiss {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var keys []*datastore.Key
	for _, b := range bytes.SplitN(i.Value, separator, limit+1)[:limit] {
		key, err := datastore.DecodeKey(string(b))
		if err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}

	return keys, nil
}
