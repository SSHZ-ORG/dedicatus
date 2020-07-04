package tgapi

import (
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/net/context"
)

type contextKey int

const (
	tgKey contextKey = iota
)

type tgContextData struct {
	bot                    *tgbotapi.BotAPI
	user                   *tgbotapi.User
	isAdmin, isContributor bool
}

func NewContext(ctx context.Context, user *tgbotapi.User) context.Context {
	c := GetConfig(ctx)
	bot := NewTgBotNoCheck(ctx)
	return context.WithValue(ctx, tgKey, tgContextData{
		bot:           bot,
		user:          user,
		isAdmin:       c.IsAdmin(user.ID),
		isContributor: c.IsContributor(user.ID),
	})
}

func fromContext(ctx context.Context) tgContextData {
	return ctx.Value(tgKey).(tgContextData)
}

func BotFromContext(ctx context.Context) *tgbotapi.BotAPI {
	return fromContext(ctx).bot
}

func UserFromContext(ctx context.Context) *tgbotapi.User {
	return fromContext(ctx).user
}

func IsAdmin(ctx context.Context) bool {
	return fromContext(ctx).isAdmin
}

func IsContributor(ctx context.Context) bool {
	return fromContext(ctx).isContributor
}
