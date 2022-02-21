package handlers

import (
	"context"

	"github.com/SSHZ-ORG/dedicatus/models"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleChosenInlineResult(ctx context.Context, update tgbotapi.Update) error {
	resultID := update.ChosenInlineResult.ResultID
	return models.IncrementUsageCounter(ctx, resultID)
}
