package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/SSHZ-ORG/dedicatus/config"
	"github.com/SSHZ-ORG/dedicatus/handlers"
	"github.com/SSHZ-ORG/dedicatus/models"
	"github.com/SSHZ-ORG/dedicatus/utils"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/gorilla/mux"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

func main() {
	r := mux.NewRouter()
	for _, token := range config.TgTokens {
		r.HandleFunc(utils.TgWebhookPath(token), generateWebhook(token))
	}
	r.HandleFunc("/admin/register", register)

	http.Handle("/", r)
	appengine.Main()
}

func register(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	err := models.CreateConfig(ctx)
	if err != nil {
		log.Errorf(ctx, "%v", err)
		http.Error(w, err.Error(), 500)
		return
	}

	var responses [][]byte

	for _, token := range config.TgTokens {
		bot, err := utils.NewTgBot(ctx, token)
		if err != nil {
			log.Errorf(ctx, "%v", err)
			http.Error(w, err.Error(), 500)
			return
		}

		response, err := utils.RegisterWebhook(ctx, bot)
		if err != nil {
			log.Errorf(ctx, "%v", err)
			http.Error(w, err.Error(), 500)
			return
		}
		responses = append(responses, response.Result)
	}

	_, _ = w.Write(bytes.Join(responses, []byte("\n")))
}

func generateWebhook(token string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := appengine.NewContext(r)

		b, _ := ioutil.ReadAll(r.Body)

		var update tgbotapi.Update
		err := json.Unmarshal(b, &update)
		if err != nil {
			log.Errorf(ctx, "%v", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
		}

		log.Infof(ctx, "%v", string(b))

		bot := utils.NewTgBotNoCheck(ctx, token)

		if update.Message != nil {
			err = handlers.HandleMessage(ctx, update, bot)
			// Internal errors from this handler should not be retried - log and tell TG we are good.
			if err != nil {
				log.Errorf(ctx, "%v", err)
				return
			}
		}

		if update.InlineQuery != nil {
			err = handlers.HandleInlineQuery(ctx, update, bot)
		}

		if update.ChosenInlineResult != nil {
			err = handlers.HandleChosenInlineResult(ctx, update, bot)
		}

		if err != nil {
			log.Errorf(ctx, "%v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}
