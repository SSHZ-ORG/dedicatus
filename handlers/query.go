package handlers

import (
	"strings"

	"github.com/SSHZ-ORG/dedicatus/models"
	"github.com/SSHZ-ORG/dedicatus/utils"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"gopkg.in/telegram-bot-api.v4"
)

func constructInlineResults(inventories []*models.Inventory) []interface{} {
	var results []interface{}
	for _, i := range inventories {
		results = append(results, utils.MakeInlineQueryResult(i.FileID, i.FileType))
	}
	return results
}

func HandleInlineQuery(ctx context.Context, update tgbotapi.Update, bot *tgbotapi.BotAPI) error {
	query := update.InlineQuery

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
			return err
		}
		_, err = bot.AnswerInlineQuery(tgbotapi.InlineConfig{
			InlineQueryID: query.ID,
			Results:       constructInlineResults(inventories),
		})
		return err
	}

	pKeysChan := make(chan *datastore.Key, len(qs))
	errsChan := make(chan error, len(qs))

	for _, q := range qs {
		go func(q string) {
			pKey, err := models.TryFindPersonalityWithKG(ctx, q)
			pKeysChan <- pKey
			errsChan <- err
		}(q)
	}

	var pKeys []*datastore.Key
	for i := 0; i < len(qs); i++ {
		err := <-errsChan
		if err != nil {
			return err
		}

		pKey := <-pKeysChan
		if pKey == nil {
			_, err = bot.AnswerInlineQuery(tgbotapi.InlineConfig{
				InlineQueryID: query.ID,
			})
			return err
		}
		pKeys = append(pKeys, pKey)
	}

	inventories, nextCursor, err := models.FindInventories(ctx, pKeys, query.Offset)
	if err != nil {
		return err
	}

	inlineConfig := tgbotapi.InlineConfig{
		InlineQueryID: query.ID,
		Results:       constructInlineResults(inventories),
		NextOffset:    nextCursor,
	}

	_, err = bot.AnswerInlineQuery(inlineConfig)
	return err
}
