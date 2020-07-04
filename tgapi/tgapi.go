package tgapi

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/SSHZ-ORG/dedicatus/config"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
)

func newTgBotNoCheck(ctx context.Context) *tgbotapi.BotAPI {
	bot := &tgbotapi.BotAPI{
		Token:  config.TgToken,
		Client: urlfetch.Client(ctx),
		Buffer: 100,
	}
	bot.SetAPIEndpoint(tgbotapi.APIEndpoint)
	return bot
}

func RegisterWebhook(ctx context.Context, bot *tgbotapi.BotAPI) (tgbotapi.APIResponse, error) {
	c := tgbotapi.NewWebhook(fmt.Sprintf("https://%s%s", appengine.DefaultVersionHostname(ctx), TgWebhookPath(bot.Token)))
	return bot.Request(c)
}

func TgWebhookPath(token string) string {
	return fmt.Sprintf("/webhook/%x", md5.Sum([]byte(token)))
}

func FetchFileInfo(ctx context.Context, fileID string) (*tgbotapi.File, []byte, error) {
	bot := BotFromContext(ctx)
	file, err := bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		return nil, nil, err
	}

	res, err := http.Get(file.Link(config.TgToken))
	if err != nil {
		return nil, nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, nil, errors.New("HTTP Status: " + res.Status)
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, nil, err
	}

	return &file, b, nil
}
