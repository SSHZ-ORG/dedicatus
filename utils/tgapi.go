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

func NewTgBotNoCheck(ctx context.Context) *tgbotapi.BotAPI {
	return &tgbotapi.BotAPI{
		Token:  dedicatus.TgToken,
		Client: urlfetch.Client(ctx),
		Buffer: 100,
	}
}

func RegisterWebhook(ctx context.Context, bot *tgbotapi.BotAPI) (tgbotapi.APIResponse, error) {
	return bot.SetWebhook(tgbotapi.NewWebhook(fmt.Sprintf("https://%s%s", appengine.DefaultVersionHostname(ctx), TgWebhookPath(bot.Token))))
}
