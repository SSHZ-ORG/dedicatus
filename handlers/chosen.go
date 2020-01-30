package handlers

import (
	"github.com/SSHZ-ORG/dedicatus/models"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"golang.org/x/net/context"
)

func HandleChosenInlineResult(ctx context.Context, update tgbotapi.Update, bot *tgbotapi.BotAPI) error {
	resultID := update.ChosenInlineResult.ResultID
	return models.IncrementUsageCounter(ctx, resultID)
}
