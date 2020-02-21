package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/SSHZ-ORG/dedicatus/config"
	"github.com/SSHZ-ORG/dedicatus/handlers"
	"github.com/SSHZ-ORG/dedicatus/models"
	"github.com/SSHZ-ORG/dedicatus/paths"
	"github.com/SSHZ-ORG/dedicatus/scheduler"
	"github.com/SSHZ-ORG/dedicatus/tgapi"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/gorilla/mux"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc(tgapi.TgWebhookPath(config.TgToken), webhook)
	r.HandleFunc("/admin/register", register)
	r.HandleFunc(paths.UpdateFileMetadata, updateFileMetadata)
	r.HandleFunc(paths.QueueUpdateFileMetadata, queueUpdateFileMetadata)

	http.Handle("/", r)
	appengine.Main()
}

func register(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	err := models.CreateConfig(ctx)
	if err != nil {
		log.Errorf(ctx, "%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	bot, err := tgapi.NewTgBot(ctx)
	if err != nil {
		log.Errorf(ctx, "%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response, err := tgapi.RegisterWebhook(ctx, bot)
	if err != nil {
		log.Errorf(ctx, "%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, _ = w.Write(response.Result)
}

func updateFileMetadata(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	id := r.FormValue("id")
	if id == "" {
		http.Error(w, "Missing arg id", http.StatusBadRequest)
		return
	}

	err := models.UpdateFileMetadata(ctx, id)
	if err != nil {
		log.Errorf(ctx, "models.UpdateFileMetadata: %+v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func queueUpdateFileMetadata(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	ids, err := models.AllInventoriesStorageKeys(ctx)
	if err != nil {
		log.Errorf(ctx, "models.AllInventoriesStorageKeys: %+v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	err = scheduler.ScheduleUpdateFileMetadata(ctx, ids)
	if err != nil {
		log.Errorf(ctx, "scheduler.ScheduleUpdateFileMetadata: %+v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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

	bot := tgapi.NewTgBotNoCheck(ctx)

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
