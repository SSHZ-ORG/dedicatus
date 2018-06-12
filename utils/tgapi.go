package utils

import (
	"fmt"

	"github.com/SSHZ-ORG/dedicatus"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
	"gopkg.in/telegram-bot-api.v4"
)

func NewTgBot(ctx context.Context) (*tgbotapi.BotAPI, error) {
	return tgbotapi.NewBotAPIWithClient(dedicatus.TgToken, urlfetch.Client(ctx))
}

func RegisterWebhook(ctx context.Context, bot *tgbotapi.BotAPI) (tgbotapi.APIResponse, error) {
	return bot.SetWebhook(tgbotapi.NewWebhook(fmt.Sprintf("https://%s/webhook/%s", appengine.DefaultVersionHostname(ctx), bot.Token)))
}
