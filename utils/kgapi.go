package utils

import (
	"encoding/json"
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

// this returns the `result` node of the found entity.
func sendKGEntityQuery(ctx context.Context, query string) (map[string]interface{}, error) {
	s, err := kgsearch.New(&http.Client{
		Transport: &transport.APIKey{
			Key:       dedicatus.KGAPIKey,
			Transport: urlfetch.Client(ctx).Transport,
		},
	})
	if err != nil {
		return nil, err
	}

	resp, err := kgsearch.NewEntitiesService(s).Search().Limit(1).Languages("ja", "zh").Query(query).Types("Person").Do()
	if err != nil {
		return nil, err
	}

	if len(resp.ItemListElement) > 0 {
		return resp.ItemListElement[0].(map[string]interface{})["result"].(map[string]interface{}), nil
	}
	return nil, nil
}

func tryFindKGEntityInternal(ctx context.Context, query string) (string, error) {
	result, err := sendKGEntityQuery(ctx, query)
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(result["@id"].(string), "kg:"), nil
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

func GetKGQueryResult(ctx context.Context, query string) (string, error) {
	// This bypasses memcache
	result, err := sendKGEntityQuery(ctx, query)
	if err != nil {
		return "", err
	}
	encoded, err := json.MarshalIndent(result, "", "    ")
	return string(encoded), err
}
