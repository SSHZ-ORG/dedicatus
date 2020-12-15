package handlers

import (
	"github.com/SSHZ-ORG/dedicatus/dctx"
	"github.com/SSHZ-ORG/dedicatus/models"
	"github.com/SSHZ-ORG/dedicatus/tgapi"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/net/context"
)

func constructInlineResults(inventories []*models.Inventory) []interface{} {
	var results []interface{}
	for _, i := range inventories {
		results = append(results, tgapi.MakeInlineQueryResult(i.FileUniqueID, i.FileID, i.FileType))
	}
	return results
}

func HandleInlineQuery(ctx context.Context, query *tgbotapi.InlineQuery) (*tgbotapi.InlineConfig, error) {
	inventories, nextCursor, err := models.QueryInventories(ctx, query.Query, query.Offset, query.ID, 50)
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
