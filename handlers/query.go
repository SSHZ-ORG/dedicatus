package handlers

import (
	"strings"

	"github.com/SSHZ-ORG/dedicatus/models"
	"github.com/SSHZ-ORG/dedicatus/utils"
	"golang.org/x/net/context"
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

	q := strings.TrimSpace(query.Query)
	if len(q) == 0 {
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

	pKey, err := models.TryFindPersonalityWithKG(ctx, q)
	if err != nil {
		return err
	}

	if pKey == nil {
		_, err = bot.AnswerInlineQuery(tgbotapi.InlineConfig{
			InlineQueryID: query.ID,
		})
		return err
	}

	inventories, nextCursor, err := models.FindInventories(ctx, pKey, query.Offset)
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
