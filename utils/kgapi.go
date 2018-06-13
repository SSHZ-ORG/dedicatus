package utils

import (
	"net/http"
	"strings"
	"time"

	"github.com/SSHZ-ORG/dedicatus"
	"golang.org/x/net/context"
	"google.golang.org/api/googleapi/transport"
	"google.golang.org/api/kgsearch/v1"
	"google.golang.org/appengine/memcache"
	"google.golang.org/appengine/urlfetch"
)

const kgMemcacheKey = "KG1:"

func tryFindKGEntityInternal(ctx context.Context, query string) (string, error) {
	s, err := kgsearch.New(&http.Client{
		Transport: &transport.APIKey{
			Key:       dedicatus.KGAPIKey,
			Transport: urlfetch.Client(ctx).Transport,
		},
	})
	if err != nil {
		return "", err
	}

	resp, err := kgsearch.NewEntitiesService(s).Search().Limit(1).Languages("ja", "zh").Query(query).Types("Person").Do()
	if err != nil {
		return "", err
	}

	if len(resp.ItemListElement) > 0 {
		respID := resp.ItemListElement[0].(map[string]interface{})["result"].(map[string]interface{})["@id"].(string)
		return strings.TrimPrefix(respID, "kg:"), nil
	}
	return "", nil
}

func getKGMemcacheKey(query string) string {
	return kgMemcacheKey + query
}

func getKGMemcache(ctx context.Context, query string) *string {
	item, err := memcache.Get(ctx, getKGMemcacheKey(query))
	if err == nil {
		s := string(item.Value)
		return &s
	}
	return nil
}

func setKGMemcache(ctx context.Context, query, result string) {
	memcache.Set(ctx, &memcache.Item{
		Key:        getKGMemcacheKey(query),
		Value:      []byte(result),
		Expiration: 4 * time.Hour,
	})
}

func TryFindKGEntity(ctx context.Context, query string) (string, error) {
	resultFromMemcache := getKGMemcache(ctx, query)
	if resultFromMemcache != nil {
		return *resultFromMemcache, nil
	}

	result, err := tryFindKGEntityInternal(ctx, query)
	if err != nil {
		return "", err
	}

	setKGMemcache(ctx, query, result)
	return result, nil
}
