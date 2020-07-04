package tgapi

import (
	"github.com/SSHZ-ORG/dedicatus/config"
	"github.com/SSHZ-ORG/dedicatus/utils"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/qedus/nds"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

type Config struct {
	Admins       []int
	Contributors []int
}

const (
	entityKind = "Config"
	stringKey  = "necessarius"
)

func configKey(ctx context.Context) *datastore.Key {
	return datastore.NewKey(ctx, entityKind, stringKey, 0, nil)
}

func GetConfig(ctx context.Context) Config {
	c := Config{}
	_ = nds.Get(ctx, configKey(ctx), &c)
	return c
}

func SetConfig(ctx context.Context, c Config) error {
	_, err := nds.Put(ctx, configKey(ctx), &c)
	return err
}

func CreateConfig(ctx context.Context) error {
	c := GetConfig(ctx)

	if !c.IsAdmin(config.InitAdminID) {
		c.Admins = append(c.Admins, config.InitAdminID)
	}

	if !c.IsContributor(config.InitAdminID) {
		c.Contributors = append(c.Contributors, config.InitAdminID)
	}

	return SetConfig(ctx, c)
}

func (c Config) IsAdmin(userID int) bool {
	return utils.Contains(c.Admins, userID)
}

func (c Config) IsContributor(userID int) bool {
	return utils.Contains(c.Contributors, userID)
}

type contextKey int

const (
	userKey contextKey = iota
)

type userContextData struct {
	user                   *tgbotapi.User
	isAdmin, isContributor bool
}

func NewContext(ctx context.Context, user *tgbotapi.User) context.Context {
	c := GetConfig(ctx)
	return context.WithValue(ctx, userKey, userContextData{
		user:          user,
		isAdmin:       c.IsAdmin(user.ID),
		isContributor: c.IsContributor(user.ID),
	})
}
