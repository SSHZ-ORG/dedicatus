package utils

import (
	"encoding/json"
	"strings"
	"time"
	"unicode"

	"github.com/SSHZ-ORG/dedicatus/config"
	"golang.org/x/net/context"
	"google.golang.org/api/kgsearch/v1"
	"google.golang.org/api/option"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/memcache"
)

const kgMemcacheKey = "KG1:"

// this returns the `result` node of the found entity.
func sendKGEntityQuery(ctx context.Context, query string) (map[string]interface{}, error) {
	s, err := kgsearch.NewService(ctx, option.WithAPIKey(config.KGAPIKey))
	if err != nil {
		return nil, err
	}

	req := kgsearch.NewEntitiesService(s).Search().Query(query).Languages("ja", "zh").Types("Person").Limit(1)
	resp, err := req.Do()
	if err != nil {
		return nil, err
	}

	if len(resp.ItemListElement) > 0 {
		return resp.ItemListElement[0].(map[string]interface{})["result"].(map[string]interface{}), nil
	}
	return nil, nil
}

func getKGEntityID(result map[string]interface{}) string {
	if result == nil {
		return ""
	}
	return strings.TrimPrefix(result["@id"].(string), "kg:")
}

func tryFindKGEntityInternal(ctx context.Context, query string) (string, error) {
	result, err := sendKGEntityQuery(ctx, query)
	if err != nil {
		return "", err
	}
	return getKGEntityID(result), nil
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
	_ = memcache.Set(ctx, &memcache.Item{
		Key:        getKGMemcacheKey(query),
		Value:      []byte(result),
		Expiration: 4 * time.Hour,
	})
}

func TryFindKGEntity(ctx context.Context, query string) string {
	resultFromMemcache := getKGMemcache(ctx, query)
	if resultFromMemcache != nil {
		return *resultFromMemcache
	}

	result, err := tryFindKGEntityInternal(ctx, query)
	if err != nil {
		// Don't fail the query, just log and return empty result.
		log.Warningf(ctx, "tryFindKGEntityInternal: %v", err)
		return ""
	}

	setKGMemcache(ctx, query, result)
	return result
}

// Returns (RawJSON, ID, JAName, error)
func GetKGQueryResult(ctx context.Context, query string) (string, string, string, error) {
	// This bypasses memcache
	result, err := sendKGEntityQuery(ctx, query)
	if err != nil {
		return "", "", "", err
	}

	cleanDetailedDescription(ctx, result)
	encoded, err := json.MarshalIndent(result, "", "  ")
	return string(encoded), getKGEntityID(result), findJaName(ctx, result), err
}

// Remove the auto-translated versions of detailedDescription. Operates as side effect on input map.
func cleanDetailedDescription(ctx context.Context, result map[string]interface{}) {
	defer func() {
		if v := recover(); v != nil {
			// Some type assertion failed. Don't care, just log.
			log.Warningf(ctx, "cleanDetailedDescription: %v", v)
		}
	}()

	var o []map[string]interface{}
	for _, i := range result["detailedDescription"].([]interface{}) {
		m := i.(map[string]interface{})
		if m["inLanguage"].(string) != "en" && strings.HasPrefix(m["url"].(string), "https://en.wikipedia.org/") {
			continue
		}
		o = append(o, m)
	}
	result["detailedDescription"] = o
}

func findJaName(ctx context.Context, result map[string]interface{}) string {
	defer func() {
		if v := recover(); v != nil {
			// Some type assertion failed. Don't care, just log.
			log.Warningf(ctx, "cleanDetailedDescription: %v", v)
		}
	}()

	for _, i := range result["name"].([]interface{}) {
		m := i.(map[string]interface{})
		if m["@language"].(string) == "ja" {
			name := m["@value"].(string)
			return strings.Map(func(r rune) rune {
				if unicode.IsSpace(r) {
					return -1
				}
				return r
			}, name)
		}
	}
	return ""
}
