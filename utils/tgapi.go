package utils

import (
	"fmt"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
)

func NewTgBot(ctx context.Context, token string) (*tgbotapi.BotAPI, error) {
	return tgbotapi.NewBotAPIWithClient(token, urlfetch.Client(ctx))
}

func NewTgBotNoCheck(ctx context.Context, token string) *tgbotapi.BotAPI {
	return &tgbotapi.BotAPI{
		Token:  token,
		Client: urlfetch.Client(ctx),
		Buffer: 100,
	}
}

func RegisterWebhook(ctx context.Context, bot *tgbotapi.BotAPI) (tgbotapi.APIResponse, error) {
	return bot.SetWebhook(tgbotapi.NewWebhook(fmt.Sprintf("https://%s%s", appengine.DefaultVersionHostname(ctx), TgWebhookPath(bot.Token))))
}
