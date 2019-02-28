package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/SSHZ-ORG/dedicatus/config"
	"github.com/SSHZ-ORG/dedicatus/handlers"
	"github.com/SSHZ-ORG/dedicatus/models"
	"github.com/SSHZ-ORG/dedicatus/utils"
	"github.com/gorilla/mux"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"gopkg.in/telegram-bot-api.v4"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc(utils.TgWebhookPath(config.TgToken), webhook)
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

	bot, err := utils.NewTgBot(ctx)
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

	w.Write(response.Result)
}

func webhook(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	bytes, _ := ioutil.ReadAll(r.Body)

	var update tgbotapi.Update
	err := json.Unmarshal(bytes, &update)
	if err != nil {
		log.Errorf(ctx, "%v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
	}

	log.Infof(ctx, "%v", string(bytes))

	bot := utils.NewTgBotNoCheck(ctx)

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
