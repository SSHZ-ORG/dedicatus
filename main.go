package main

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"github.com/SSHZ-ORG/dedicatus/config"
	"github.com/SSHZ-ORG/dedicatus/dctx"
	"github.com/SSHZ-ORG/dedicatus/handlers"
	"github.com/SSHZ-ORG/dedicatus/handlers/messages"
	"github.com/SSHZ-ORG/dedicatus/models"
	"github.com/SSHZ-ORG/dedicatus/paths"
	"github.com/SSHZ-ORG/dedicatus/scheduler"
	"github.com/SSHZ-ORG/dedicatus/scheduler/metadatamode"
	"github.com/SSHZ-ORG/dedicatus/tgapi"
	"github.com/SSHZ-ORG/dedicatus/twapi"
	"github.com/SSHZ-ORG/dedicatus/webui"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/julienschmidt/httprouter"
	"google.golang.org/appengine/v2"
	"google.golang.org/appengine/v2/log"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	r := httprouter.New()

	r.POST(tgapi.TgWebhookPath(config.TgToken), webhook)
	r.GET(paths.RegisterWebhook, registerWebhook)
	r.POST(paths.UpdateFileMetadata, updateFileMetadata)
	r.GET(paths.QueueUpdateFileMetadata, queueUpdateFileMetadata)
	r.GET(paths.RotateReservoir, rotateReservoir)
	r.GET(paths.PostTweet, postTweet)

	r.GET(paths.WebUI, webui.Handler)

	http.Handle("/", r)
	appengine.Main()
}

func registerWebhook(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx := dctx.NewContext(r)

	bot := tgapi.BotFromContext(ctx)

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

func updateFileMetadata(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx := dctx.NewContext(r)

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

func queueUpdateFileMetadata(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx := dctx.NewContext(r)

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

func rotateReservoir(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx := dctx.NewContext(r)

	err := models.RotateReservoir(ctx)
	if err != nil {
		log.Errorf(ctx, "models.RotateReservoir: %+v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func postTweet(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx := dctx.NewContext(r)
	if _, err := twapi.SendInventoryToTwitter(ctx, r.FormValue("id")); err != nil {
		log.Errorf(ctx, "twapi.SendInventoryToTwitter: %+v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func webhook(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx := dctx.NewContext(r)

	bytes, _ := ioutil.ReadAll(r.Body)

	var update tgbotapi.Update
	err := json.Unmarshal(bytes, &update)
	if err != nil {
		log.Errorf(ctx, "%v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	log.Debugf(ctx, "%v", string(bytes))

	var response tgbotapi.Chattable

	if update.Message != nil {
		dctx.AttachUserInSession(ctx, update.Message.From)
		response, err = messages.HandleMessage(ctx, update.Message)
	}

	if update.InlineQuery != nil {
		dctx.AttachUserInSession(ctx, update.InlineQuery.From)
		response, err = handlers.HandleInlineQuery(ctx, update.InlineQuery)
	}

	if update.ChosenInlineResult != nil {
		dctx.AttachUserInSession(ctx, update.ChosenInlineResult.From)
		err = handlers.HandleChosenInlineResult(ctx, update)
	}

	if response != nil {
		err = tgbotapi.WriteToHTTPResponse(w, response)
	}

	if err != nil {
		log.Errorf(ctx, "%v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
