package handlers

import (
	"strings"

	"github.com/SSHZ-ORG/dedicatus/models"
	"github.com/SSHZ-ORG/dedicatus/utils"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

func constructInlineResults(inventories []*models.Inventory) []interface{} {
	var results []interface{}
	for _, i := range inventories {
		results = append(results, utils.MakeInlineQueryResult(i.FileID, i.FileType))
	}
	return results
}

func HandleInlineQuery(ctx context.Context, update tgbotapi.Update, bot *tgbotapi.BotAPI) error {
	response, err := prepareResponse(ctx, update.InlineQuery)
	if err != nil {
		return err
	}

	log.Debugf(ctx, "AnswerInlineQuery: %+v", *response)

	_, err = bot.AnswerInlineQuery(*response)
	return err
}

func prepareResponse(ctx context.Context, query *tgbotapi.InlineQuery) (*tgbotapi.InlineConfig, error) {
	rawQs := strings.Split(strings.TrimSpace(query.Query), " ")
	qs := rawQs[:0]
	for _, e := range rawQs {
		if e != "" {
			qs = append(qs, e)
		}
	}

	if len(qs) == 0 {
		inventories, err := models.GloballyLastUsedInventories(ctx)
		if err != nil {
			return nil, err
		}
		return &tgbotapi.InlineConfig{
			InlineQueryID: query.ID,
			Results:       constructInlineResults(inventories),
		}, nil
	}

	pKeysChan := make(chan []*datastore.Key, len(qs))
	errsChan := make(chan error, len(qs))

	for _, q := range qs {
		go func(q string) {
			pKeys, err := models.TryFindPersonalitiesWithKG(ctx, q)
			pKeysChan <- pKeys
			errsChan <- err
		}(q)
	}

	var flattenKeys []*datastore.Key
	for i := 0; i < len(qs); i++ {
		err := <-errsChan
		if err != nil {
			return nil, err
		}

		pKeys := <-pKeysChan
		if len(pKeys) == 0 {
			return &tgbotapi.InlineConfig{
				InlineQueryID: query.ID,
			}, nil
		}
		flattenKeys = append(flattenKeys, pKeys...)
	}

	inventories, nextCursor, err := models.FindInventories(ctx, flattenKeys, query.Offset)
	if err != nil {
		return nil, err
	}

	return &tgbotapi.InlineConfig{
		InlineQueryID: query.ID,
		Results:       constructInlineResults(inventories),
		NextOffset:    nextCursor,
	}, nil
}
