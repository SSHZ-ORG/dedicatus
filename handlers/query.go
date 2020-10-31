package handlers

import (
	"strings"
	"sync"

	"github.com/SSHZ-ORG/dedicatus/models"
	"github.com/SSHZ-ORG/dedicatus/models/sortmode"
	"github.com/SSHZ-ORG/dedicatus/tgapi"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

func constructInlineResults(inventories []*models.Inventory) []interface{} {
	var results []interface{}
	for _, i := range inventories {
		results = append(results, tgapi.MakeInlineQueryResult(i.FileUniqueID, i.FileID, i.FileType))
	}
	return results
}

func HandleInlineQuery(ctx context.Context, query *tgbotapi.InlineQuery) (*tgbotapi.InlineConfig, error) {
	qs := strings.Fields(query.Query)

	// Determine QueryMode
	queryMode := sortmode.UsageCountDesc
	if len(qs) == 0 { // Globally last used
		queryMode = sortmode.LastUsedDesc
	}
	for i, q := range qs {
		if m := sortmode.ParseQuerySortMode(q); m != sortmode.Undefined {
			// This token is valid sort mode flag.
			queryMode = m                    // Use it
			qs = append(qs[:i], qs[i+1:]...) // Delete this token so we don't use it to search personality
			break                            // If there are more tokens after us, leave them there. It will be used to search personality and fail
		}
	}

	pKeys := make([][]*datastore.Key, len(qs))
	errs := make([]error, len(qs))
	wg := sync.WaitGroup{}
	wg.Add(len(qs))

	for i := range qs {
		go func(i int) {
			defer wg.Done()
			pKeys[i], errs[i] = models.TryFindPersonalitiesWithKG(ctx, qs[i])
		}(i)
	}

	wg.Wait()

	var flattenKeys []*datastore.Key
	for i := range qs {
		if errs[i] != nil {
			return nil, errs[i]
		}

		if len(pKeys[i]) == 0 {
			// One token is neither sort mode flag nor personality. Return empty result.
			return &tgbotapi.InlineConfig{
				InlineQueryID: query.ID,
			}, nil
		}
		flattenKeys = append(flattenKeys, pKeys[i]...)
	}

	inventories, nextCursor, err := models.QueryInventories(ctx, flattenKeys, queryMode, query.Offset, query.ID)
	if err != nil {
		return nil, err
	}

	return &tgbotapi.InlineConfig{
		InlineQueryID: query.ID,
		Results:       constructInlineResults(inventories),
		NextOffset:    nextCursor,
		CacheTime:     120,
	}, nil
}
