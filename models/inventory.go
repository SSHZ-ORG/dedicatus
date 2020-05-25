package models

import (
	"crypto/md5"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/SSHZ-ORG/dedicatus/models/cursor"
	"github.com/SSHZ-ORG/dedicatus/models/reservoir"
	"github.com/SSHZ-ORG/dedicatus/models/sortmode"
	"github.com/SSHZ-ORG/dedicatus/scheduler"
	"github.com/SSHZ-ORG/dedicatus/scheduler/metadatamode"
	"github.com/SSHZ-ORG/dedicatus/tgapi"
	"github.com/SSHZ-ORG/dedicatus/utils"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/qedus/nds"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

const maxItems = 50

var (
	ErrorOnlyAdminCanUpdateInventory = errors.New("Only admins can update an existing GIF.")
	ErrorHashConflict                = errors.New("Hash conflict")
)

type Inventory struct {
	FileUniqueID string
	FileID       string
	FileName     string
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

	return fmt.Sprintf("%s\nUniqueID: %s\n%x (%d bytes)\n[%s]", i.FileName, i.FileUniqueID, i.MD5Sum, i.FileSize, strings.Join(pns, ", ")), nil
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
func TryGetInventoryByTgAnimation(ctx context.Context, animation *tgbotapi.ChatAnimation) (*Inventory, error) {
	i, err := GetInventory(ctx, animation.FileUniqueID)
	if err == nil {
		return i, nil
	}
	if err != datastore.ErrNoSuchEntity {
		return nil, err
	}

	return getInventoryByFile(ctx, animation.FileID, animation.FileSize)
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
		return nil, ErrorHashConflict
	}

	i := new(Inventory)
	err = nds.Get(ctx, keys[0], i)
	return i, err
}

func CreateOrUpdateInventory(ctx context.Context, fileID, fileUniqueID, fileName string, personality []*datastore.Key, userID int, config Config) (*Inventory, error) {
	i := new(Inventory)

	err := nds.RunInTransaction(ctx, func(ctx context.Context) error {
		key := inventoryKey(ctx, fileUniqueID)
		err := nds.Get(ctx, key, i)

		// This is an existing Inventory, only admins or original creator can update it.
		if err == nil && !(config.IsAdmin(userID) || i.Creator == userID) {
			return ErrorOnlyAdminCanUpdateInventory
		}
		if err != nil && err != datastore.ErrNoSuchEntity {
			return err
		}

		i.FileID = fileID
		i.FileUniqueID = fileUniqueID

		// For existing Inventory that we know the name, don't override.
		if i.FileName == "" {
			i.FileName = fileName
		}

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
			if err := scheduler.ScheduleUpdateFileMetadata(ctx, []string{fileUniqueID}, metadatamode.Default); err != nil {
				return err
			}
		}
		return nil
	}, &datastore.TransactionOptions{})

	return i, err
}

func queryInventoryKeys(ctx context.Context, personalities []*datastore.Key, sortMode sortmode.SortMode, lastCursor, queryID string) ([]*datastore.Key, string, error) {
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
		if len(personalities) == 0 {
			// Global random. Use reservoir.
			keys, err := reservoir.ReadReservoir(ctx, maxItems)
			return keys, "", err
		}
		// Not implemented. Let it use whatever natural ordering for now.
	}

	var offset int
	q, offset = cursor.Offset(ctx, q, lastCursor)
	q = q.Limit(maxItems)

	var keys []*datastore.Key
	it := q.Run(ctx)
	key, err := it.Next(nil)
	for err == nil {
		keys = append(keys, key)
		key, err = it.Next(nil)
	}
	if err != datastore.Done {
		return nil, "", err
	}

	if len(keys) == 0 {
		return nil, "", nil
	}

	newCursor := ""
	if len(keys) == maxItems {
		c, _ := it.Cursor()
		newCursor = cursor.Store(ctx, c, queryID, offset+maxItems)
	}

	return keys, newCursor, nil
}

func QueryInventories(ctx context.Context, personalities []*datastore.Key, sortMode sortmode.SortMode, lastCursor, queryID string) ([]*Inventory, string, error) {
	keys, newCursor, err := queryInventoryKeys(ctx, personalities, sortMode, lastCursor, queryID)

	g := make([]*Inventory, len(keys))
	err = nds.GetMulti(ctx, keys, g)
	if err != nil {
		if tryFlattenDatastoreNoSuchEntityMultiError(err) != datastore.ErrNoSuchEntity {
			return nil, "", err
		}
	}

	var is []*Inventory
	for _, i := range g {
		if i != nil {
			is = append(is, i)
		}
	}

	return is, newCursor, nil
}

func allInventories(ctx context.Context) ([]*datastore.Key, error) {
	return datastore.NewQuery(inventoryEntityKind).KeysOnly().GetAll(ctx, nil)

}

func AllInventoriesStorageKeys(ctx context.Context) ([]string, error) {
	keys, err := allInventories(ctx)
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

func ReplaceFileID(ctx context.Context, oldFileUniqueID, newFileID, newFileUniqueID, newFileName string) (*Inventory, error) {
	i := new(Inventory)

	err := nds.RunInTransaction(ctx, func(ctx context.Context) error {
		oldKey := inventoryKey(ctx, oldFileUniqueID)
		if err := nds.Get(ctx, oldKey, i); err != nil {
			return err
		}

		i.FileID = newFileID
		i.FileUniqueID = newFileUniqueID
		i.FileName = newFileName

		i.MD5Sum = nil
		i.FileSize = 0

		if err := nds.Delete(ctx, oldKey); err != nil {
			return err
		}
		if _, err := nds.Put(ctx, inventoryKey(ctx, newFileUniqueID), i); err != nil {
			return err
		}

		return scheduler.ScheduleUpdateFileMetadata(ctx, []string{newFileUniqueID}, metadatamode.Default)
	}, &datastore.TransactionOptions{XG: true})

	return i, err
}

func OverrideFileName(ctx context.Context, fileUniqueID, fileName string) error {
	i := new(Inventory)
	key := inventoryKey(ctx, fileUniqueID)

	return nds.RunInTransaction(ctx, func(ctx context.Context) error {
		if err := nds.Get(ctx, key, i); err != nil {
			return err
		}

		i.FileName = fileName

		_, err := nds.Put(ctx, key, i)
		return err
	}, &datastore.TransactionOptions{})
}

func RotateReservoir(ctx context.Context) error {
	if reservoir.TryRotateReservoir(ctx, maxItems) {
		return nil
	}

	keys, err := allInventories(ctx)
	if err != nil {
		return err
	}
	return reservoir.RefillReservoir(ctx, keys)
}
