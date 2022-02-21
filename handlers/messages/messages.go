package messages

import (
	"context"
	"fmt"
	"strings"

	"github.com/SSHZ-ORG/dedicatus/dctx"
	"github.com/SSHZ-ORG/dedicatus/dctx/protoconf/pb"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"google.golang.org/appengine/v2/log"
)

const (
	errorMessageNotAdmin       = "Only admins can do this, sorry."
	errorMessageNotContributor = "Only contributors can do this, sorry."
)

func makeReplyMessage(message *tgbotapi.Message, reply string) *tgbotapi.MessageConfig {
	c := tgbotapi.NewMessage(message.Chat.ID, reply)
	c.ReplyToMessageID = message.MessageID
	c.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	c.DisableWebPagePreview = true
	return &c
}

func HandleMessage(ctx context.Context, message *tgbotapi.Message) (tgbotapi.Chattable, error) {
	var response tgbotapi.Chattable
	var err error

	if message.Animation != nil || (message.ReplyToMessage != nil && message.ReplyToMessage.Animation != nil) {
		response, err = handleAnimation(ctx, message)
	} else if message.Video != nil || (message.ReplyToMessage != nil && message.ReplyToMessage.Video != nil) {
		response, err = handleVideo(ctx, message)
	} else if message.Contact != nil {
		response, err = handleContact(ctx, message)
	} else if strings.HasPrefix(message.Text, "/") {
		args := strings.Fields(message.Text)

		if handler, ok := commandMap[args[0]]; ok {
			res := ""
			res, err = handler(ctx, args)
			response = makeReplyMessage(message, res)
		} else if handler, ok := complexCommandMap[args[0]]; ok {
			response, err = handler(ctx, args, message)
		} else {
			response = makeReplyMessage(message, "Command not recognized.")
		}
	}

	if err != nil {
		// Don't cause retry. Log and respond that we have an internal error.
		log.Errorf(ctx, "%v", err)
		response = makeReplyMessage(message, "Your action triggered an internal server error.")
	}

	return response, nil
}

func handleContact(ctx context.Context, message *tgbotapi.Message) (tgbotapi.Chattable, error) {
	if !dctx.IsAdmin(ctx) {
		return makeReplyMessage(message, errorMessageNotAdmin), nil
	}

	uid := message.Contact.UserID
	if uid == 0 {
		return nil, nil
	}

	a := dctx.ProtoconfFromContext(ctx).GetAuthConfig()
	reply := makeReplyMessage(message, fmt.Sprintf("User %d\nType: %v", uid, a.GetUsers()[int64(uid)].String()))

	keyboard := tgbotapi.NewOneTimeReplyKeyboard()
	for _, t := range []pb.AuthConfig_UserType{pb.AuthConfig_USER, pb.AuthConfig_CONTRIBUTOR, pb.AuthConfig_ADMIN} {
		kb := []tgbotapi.KeyboardButton{tgbotapi.NewKeyboardButton(fmt.Sprintf("/c auth %d %s", uid, t))}
		keyboard.Keyboard = append(keyboard.Keyboard, kb)
	}
	keyboard.Selective = true
	reply.ReplyMarkup = keyboard

	return reply, nil
}
