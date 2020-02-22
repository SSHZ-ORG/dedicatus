package models

import (
	"crypto/md5"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/SSHZ-ORG/dedicatus/models/sortmode"
	"github.com/SSHZ-ORG/dedicatus/scheduler"
	"github.com/SSHZ-ORG/dedicatus/tgapi"
	"github.com/SSHZ-ORG/dedicatus/utils"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/qedus/nds"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

const maxItems = 50

var (
	ErrorOnlyAdminCanUpdateInventory = errors.New("Only admins can update an existing GIF.")
)

type Inventory struct {
	FileUniqueID string
	FileID       string
	FileType     string
	Personality  []*datastore.Key
	Creator      int

	UsageCount int64
	LastUsed   time.Time

	MD5Sum   datastore.ByteString
	FileSize int
}

func (i Inventory) ToString(ctx context.Context) (string, error) {
	ps := make([]*Personality, len(i.Personality))
	err := nds.GetMulti(ctx, i.Personality, ps)
	if err != nil {
		return "", err
	}

	var pns []string
	for _, p := range ps {
		pns = append(pns, p.CanonicalName)
	}

	return fmt.Sprintf("UniqueID: %s\n%x (%d bytes)\n[%s]", i.FileUniqueID, i.MD5Sum, i.FileSize, strings.Join(pns, ", ")), nil
}

func inventoryKey(ctx context.Context, fileUniqueID string) *datastore.Key {
	return datastore.NewKey(ctx, inventoryEntityKind, fileUniqueID, 0, nil)
}

func GetInventory(ctx context.Context, fileUniqueID string) (*Inventory, error) {
	i := new(Inventory)
	key := inventoryKey(ctx, fileUniqueID)
	err := nds.Get(ctx, key, i)
	return i, err
}

// If not found, returns (nil, datastore.ErrNoSuchEntity)
func TryGetInventoryByTgDocument(ctx context.Context, document *tgbotapi.Document) (*Inventory, error) {
	i, err := GetInventory(ctx, document.FileUniqueID)
	if err == nil {
		return i, nil
	}
	if err != datastore.ErrNoSuchEntity {
		return nil, err
	}

	return getInventoryByFile(ctx, document.FileID, document.FileSize)
}

// Matches the given file with known Inventories without using UniqueID. If not found returns (nil, datastore.ErrNoSuchEntity).
func getInventoryByFile(ctx context.Context, fileID string, fileSize int) (*Inventory, error) {
	count, err := datastore.NewQuery(inventoryEntityKind).Filter("FileSize =", fileSize).Count(ctx)
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, datastore.ErrNoSuchEntity
	}

	_, b, err := tgapi.FetchFileInfo(ctx, fileID)
	if err != nil {
		return nil, err
	}

	s := md5.Sum(b)
	return getInventoryByMD5(ctx, s[:])
}

// If not found, returns (nil, datastore.ErrNoSuchEntity)
func getInventoryByMD5(ctx context.Context, sum []byte) (*Inventory, error) {
	keys, err := datastore.NewQuery(inventoryEntityKind).Filter("MD5Sum =", sum[:]).KeysOnly().GetAll(ctx, nil)
	if err != nil {
		return nil, err
	}

	if len(keys) == 0 {
		return nil, datastore.ErrNoSuchEntity
	} else if len(keys) > 1 {
		log.Criticalf(ctx, "Hash conflict (%x)!", sum)
		return nil, datastore.ErrNoSuchEntity
	}

	i := new(Inventory)
	err = nds.Get(ctx, keys[0], i)
	return i, err
}

// Not migrated, don't use.
func CreateInventory(ctx context.Context, fileID string, personality []*datastore.Key, userID int, config Config) (*Inventory, error) {
	i := new(Inventory)

	err := nds.RunInTransaction(ctx, func(ctx context.Context) error {
		key := inventoryKey(ctx, fileID)
		err := nds.Get(ctx, key, i)

		// This is an existing Inventory, only admins or original creator can update it.
		if err == nil && !(config.IsAdmin(userID) || i.Creator == userID) {
			return ErrorOnlyAdminCanUpdateInventory
		}
		if err != nil && err != datastore.ErrNoSuchEntity {
			return err
		}

		i.FileID = fileID
		i.FileType = utils.FileTypeMPEG4GIF
		i.Personality = personality
		i.LastUsed = time.Now()

		if i.Creator == 0 {
			i.Creator = userID
		}

		shouldScheduleMetadataUpdate := err == datastore.ErrNoSuchEntity

		if _, err := nds.Put(ctx, key, i); err != nil {
			return err
		}

		if shouldScheduleMetadataUpdate {
			// New File. Schedule metadata update.
			if err := scheduler.ScheduleUpdateFileMetadata(ctx, []string{fileID}); err != nil {
				return err
			}
		}
		return nil
	}, &datastore.TransactionOptions{})

	return i, err
}

func FindInventories(ctx context.Context, personalities []*datastore.Key, sortMode sortmode.SortMode, lastCursor string) ([]*Inventory, string, error) {
	q := datastore.NewQuery(inventoryEntityKind).KeysOnly()

	for _, personality := range personalities {
		q = q.Filter("Personality = ", personality)
	}

	switch sortMode {
	case sortmode.UsageCountDesc:
		q = q.Order("-UsageCount")
	case sortmode.UsageCountAsc:
		q = q.Order("UsageCount")
	case sortmode.LastUsedDesc:
		q = q.Order("-LastUsed")
	case sortmode.LastUsedAsc:
		q = q.Order("LastUsed")
	case sortmode.RandomDraw:
		// Not implemented. Let it use whatever natural ordering for now.
	}

	offset, err := strconv.Atoi(lastCursor)
	if err == nil {
		q = q.Offset(offset)
	}

	q = q.Limit(maxItems)

	keys, err := q.GetAll(ctx, nil)
	if err != nil {
		return nil, "", err
	}

	if len(keys) == 0 {
		return nil, "", nil
	}

	inventories := make([]*Inventory, len(keys))
	err = nds.GetMulti(ctx, keys, inventories)
	if err != nil {
		return nil, "", err
	}

	newCursor := ""
	if len(keys) == maxItems {
		newCursor = strconv.Itoa(offset + maxItems)
	}

	return inventories, newCursor, nil
}

func AllInventoriesStorageKeys(ctx context.Context) ([]string, error) {
	keys, err := datastore.NewQuery(inventoryEntityKind).KeysOnly().GetAll(ctx, nil)
	if err != nil {
		return nil, err
	}

	var storageKeys []string
	for _, k := range keys {
		storageKeys = append(storageKeys, k.StringID())
	}

	return storageKeys, nil
}

func IncrementUsageCounter(ctx context.Context, fileUniqueID string) error {
	key := inventoryKey(ctx, fileUniqueID)

	return nds.RunInTransaction(ctx, func(ctx context.Context) error {
		i := new(Inventory)
		// Get again to avoid race condition.
		if err := nds.Get(ctx, key, i); err != nil {
			return err
		}

		i.UsageCount += 1
		i.LastUsed = time.Now()

		_, err := nds.Put(ctx, key, i)
		return err
	}, &datastore.TransactionOptions{})
}

func CountInventories(ctx context.Context, personality *datastore.Key) (int, error) {
	return datastore.NewQuery(inventoryEntityKind).KeysOnly().Filter("Personality = ", personality).Count(ctx)
}

// Not migrated, don't use.
func ReplaceFileID(ctx context.Context, oldFileID, newFileID string) (*Inventory, error) {
	i := new(Inventory)

	err := nds.RunInTransaction(ctx, func(ctx context.Context) error {
		oldKey := inventoryKey(ctx, oldFileID)
		if err := nds.Get(ctx, oldKey, i); err != nil {
			return err
		}

		i.FileID = newFileID

		i.MD5Sum = nil
		i.FileSize = 0

		if err := nds.Delete(ctx, oldKey); err != nil {
			return err
		}
		if _, err := nds.Put(ctx, inventoryKey(ctx, newFileID), i); err != nil {
			return err
		}

		return scheduler.ScheduleUpdateFileMetadata(ctx, []string{newFileID})
	}, &datastore.TransactionOptions{XG: true})

	return i, err
}

func UpdateFileMetadata(ctx context.Context, oldStorageKey string) error {
	i := new(Inventory)
	oldKey := inventoryKey(ctx, oldStorageKey)
	if err := nds.Get(ctx, oldKey, i); err != nil {
		if err == datastore.ErrNoSuchEntity {
			// Silently ignore this.
			return nil
		}
		return err
	}

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
