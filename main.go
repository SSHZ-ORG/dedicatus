package main

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"github.com/SSHZ-ORG/dedicatus/config"
	"github.com/SSHZ-ORG/dedicatus/handlers"
	"github.com/SSHZ-ORG/dedicatus/models"
	"github.com/SSHZ-ORG/dedicatus/paths"
	"github.com/SSHZ-ORG/dedicatus/scheduler"
	"github.com/SSHZ-ORG/dedicatus/scheduler/metadatamode"
	"github.com/SSHZ-ORG/dedicatus/tgapi"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/gorilla/mux"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	r := mux.NewRouter()
	r.HandleFunc(tgapi.TgWebhookPath(config.TgToken), webhook)
	r.HandleFunc(paths.RegisterWebhook, registerWebhook)
	r.HandleFunc(paths.UpdateFileMetadata, updateFileMetadata)
	r.HandleFunc(paths.QueueUpdateFileMetadata, queueUpdateFileMetadata)
	r.HandleFunc(paths.RotateReservoir, rotateReservoir)

	http.Handle("/", r)
	appengine.Main()
}

func registerWebhook(w http.ResponseWriter, r *http.Request) {
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

	if _, err := tgapi.RegisterWebhook(ctx, bot); err != nil {
		log.Errorf(ctx, "%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	commands := tgbotapi.NewSetMyCommands(
		tgbotapi.BotCommand{Command: "n", Description: "create a new Personality"},
		tgbotapi.BotCommand{Command: "s", Description: "find existing Personalities"},
		tgbotapi.BotCommand{Command: "u", Description: "edit Alias for Personalities"},
		tgbotapi.BotCommand{Command: "kg", Description: "query Knowledge Graph"})
	if _, err := bot.Request(commands); err != nil {
		log.Errorf(ctx, "%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, _ = w.Write([]byte("OK"))
}

func updateFileMetadata(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	id := r.FormValue("id")
	if id == "" {
		http.Error(w, "Missing arg id", http.StatusBadRequest)
		return
	}

	mode := metadatamode.FromString(r.FormValue("mode"))

	err := models.UpdateFileMetadata(ctx, id, mode)
	if err != nil {
		log.Errorf(ctx, "models.UpdateFileMetadata: %+v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func queueUpdateFileMetadata(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	ids, err := models.AllInventoriesStorageKeys(ctx)
	if err != nil {
		log.Errorf(ctx, "models.AllInventoriesStorageKeys: %+v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	mode := metadatamode.FromString(r.FormValue("mode"))

	err = scheduler.ScheduleUpdateFileMetadata(ctx, ids, mode)
	if err != nil {
		log.Errorf(ctx, "scheduler.ScheduleUpdateFileMetadata: %+v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func rotateReservoir(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	err := models.RotateReservoir(ctx)
	if err != nil {
		log.Errorf(ctx, "models.RotateReservoir: %+v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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
		return
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
		return
	}
}