package handlers

import (
	"github.com/SSHZ-ORG/dedicatus/models"
	"golang.org/x/net/context"
	"gopkg.in/telegram-bot-api.v4"
)

func HandleChosenInlineResult(ctx context.Context, update tgbotapi.Update, bot *tgbotapi.BotAPI) error {
	fileID := update.ChosenInlineResult.ResultID
	return models.IncrementUsageCounter(ctx, fileID)
}
