package models

import (
	"crypto/md5"
	"errors"

	"github.com/SSHZ-ORG/dedicatus/config"
	"github.com/SSHZ-ORG/dedicatus/scheduler/metadatamode"
	"github.com/SSHZ-ORG/dedicatus/tgapi"
	"github.com/SSHZ-ORG/dedicatus/utils"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/qedus/nds"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

var metadataUpdaters = map[metadatamode.MetadataMode]func(context.Context, string, *Inventory) error{
	metadatamode.Default:      updateFileMetadataDefault,
	metadatamode.ReadFileName: updateFileMetadataReadFileName,
}

func UpdateFileMetadata(ctx context.Context, oldStorageKey string, mode metadatamode.MetadataMode) error {
	i := new(Inventory)
	oldKey := inventoryKey(ctx, oldStorageKey)
	if err := nds.Get(ctx, oldKey, i); err != nil {
		if err == datastore.ErrNoSuchEntity {
			// Silently ignore this.
			return nil
		}
		return err
	}

	if updater, ok := metadataUpdaters[mode]; ok {
		return updater(ctx, oldStorageKey, i)
	}
	return errors.New("unknown MetadataMode")
}

func updateFileMetadataDefault(ctx context.Context, oldStorageKey string, i *Inventory) error {
	oldKey := inventoryKey(ctx, oldStorageKey)

	file, b, err := tgapi.FetchFileInfo(ctx, i.FileID)
	if err != nil {
		return err
	}

	newFileID := file.FileID
	fileUniqueID := file.FileUniqueID

	if newFileID != i.FileID {
		log.Infof(ctx, "Detected FileID change %s -> %s for FileUniqueID %s", oldStorageKey, newFileID, fileUniqueID)
	}

	if !appengine.IsDevAppServer() {
		if err := utils.StoreFileToGCS(ctx, fileUniqueID, b); err != nil {
			return err
		}
	}

	sum := md5.Sum(b)
	log.Infof(ctx, "File %s: %x (%d bytes)", fileUniqueID, sum, file.FileSize)

	return nds.RunInTransaction(ctx, func(ctx context.Context) error {
		// Get again so we don't race.
		if err := nds.Get(ctx, oldKey, i); err != nil {
			if err == datastore.ErrNoSuchEntity {
				// Silently ignore this.
				return nil
			}
			return err
		}

		i.FileUniqueID = fileUniqueID
		i.FileID = newFileID
		i.MD5Sum = sum[:]
		i.FileSize = file.FileSize

		if oldStorageKey != i.FileUniqueID {
			if err := nds.Delete(ctx, oldKey); err != nil {
				return err
			}
		}
		_, err := nds.Put(ctx, inventoryKey(ctx, i.FileUniqueID), i)
		return err
	}, &datastore.TransactionOptions{XG: true})
}

func updateFileMetadataReadFileName(ctx context.Context, oldStorageKey string, i *Inventory) error {
	if i.FileName != "" {
		return nil
	}

	bot := tgapi.NewTgBotNoCheck(ctx)

	// Let's spam the init admin to get FileName.
	a := tgbotapi.NewAnimationShare(config.InitAdminID, i.FileID)
	a.Caption = "Reading FileName for " + i.FileUniqueID
	m, err := bot.Send(a)
	if err != nil {
		return err
	}

	return OverrideFileName(ctx, i.FileUniqueID, m.Animation.FileName)
}
