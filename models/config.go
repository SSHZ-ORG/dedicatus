package models

import (
	"github.com/SSHZ-ORG/dedicatus/config"
	"github.com/SSHZ-ORG/dedicatus/utils"
	"github.com/qedus/nds"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

type Config struct {
	Admins       []int
	Contributors []int
}

const stringKey = "necessarius"

func configKey(ctx context.Context) *datastore.Key {
	return datastore.NewKey(ctx, configEntityKind, stringKey, 0, nil)
}

func GetConfig(ctx context.Context) Config {
	c := Config{}
	nds.Get(ctx, configKey(ctx), &c)
	return c
}

func SetConfig(ctx context.Context, c Config) error {
	_, err := nds.Put(ctx, configKey(ctx), &c)
	return err
}

func CreateConfig(ctx context.Context) error {
	c := GetConfig(ctx)

	admins := utils.NewIntSetFromSlice(c.Admins)
	admins.Add(config.InitAdminID)

	contributors := utils.NewIntSetFromSlice(c.Contributors)
	contributors.Add(config.InitAdminID)

	c.Admins = admins.ToSlice()
	c.Contributors = contributors.ToSlice()

	_, err := nds.Put(ctx, configKey(ctx), &c)
	return err
}

func (c Config) IsAdmin(userID int) bool {
	return utils.Contains(c.Admins, userID)
}

func (c Config) IsContributor(userID int) bool {
	return utils.Contains(c.Contributors, userID)
}
