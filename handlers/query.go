package handlers

import (
	"strings"
	"sync"

	"github.com/SSHZ-ORG/dedicatus/dctx"
	"github.com/SSHZ-ORG/dedicatus/models"
	"github.com/SSHZ-ORG/dedicatus/models/sortmode"
	"github.com/SSHZ-ORG/dedicatus/tgapi"
	"github.com/SSHZ-ORG/dedicatus/utils"
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

	var pqs, tqs []string

	// Determine QueryMode
	queryMode := sortmode.UsageCountDesc
	if len(qs) == 0 { // Globally last used
		queryMode = sortmode.LastUsedDesc
	}
	for _, q := range qs {
		if m := sortmode.ParseQuerySortMode(q); m != sortmode.Undefined {
			// This token is valid sort mode flag.
			queryMode = m
		} else if utils.IsTagFormatted(q) {
			// This is a Tag query.
			tqs = append(tqs, utils.TrimFirstRune(q)) // Need to remove `#`
		} else {
			// Assume this is a Personality query.
			pqs = append(pqs, q)
		}
	}

	wg := sync.WaitGroup{}
	wg.Add(len(pqs) + len(tqs))

	pKeys := make([][]*datastore.Key, len(pqs))
	pErrs := make([]error, len(pqs))
	for i := range pqs {
		go func(i int) {
			defer wg.Done()
			pKeys[i], pErrs[i] = models.TryFindPersonalitiesWithKG(ctx, pqs[i])
		}(i)
	}

	tKeys := make([]*datastore.Key, len(tqs))
	tErrs := make([]error, len(tqs))
	for i := range tqs {
		go func(i int) {
			defer wg.Done()
			tKeys[i], _, tErrs[i] = models.FindTag(ctx, tqs[i])
		}(i)
	}

	wg.Wait()

	var flattenPKeys []*datastore.Key
	for i := range pqs {
		if pErrs[i] != nil {
			return nil, pErrs[i]
		}

		if len(pKeys[i]) == 0 {
			// This token is not personality. Return empty result.
			return &tgbotapi.InlineConfig{
				InlineQueryID: query.ID,
			}, nil
		}
		flattenPKeys = append(flattenPKeys, pKeys[i]...)
	}

	for i, k := range tKeys {
		if tErrs[i] != nil {
			return nil, tErrs[i]
		}
		if k == nil {
			// This token is not a tag. Return empty result.
			return &tgbotapi.InlineConfig{
				InlineQueryID: query.ID,
			}, nil
		}
	}

	inventories, nextCursor, err := models.QueryInventories(ctx, flattenPKeys, tKeys, queryMode, query.Offset, query.ID)
	if err != nil {
		return nil, err
	}

	return &tgbotapi.InlineConfig{
		InlineQueryID: query.ID,
		Results:       constructInlineResults(inventories),
		NextOffset:    nextCursor,
		CacheTime:     int(dctx.ProtoconfFromContext(ctx).GetInlineQueryCacheTimeSec()),
	}, nil
}
