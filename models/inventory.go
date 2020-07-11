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

	TwitterMediaID string
	LastTweetTime  time.Time

	ModelVersion int
}

func (i Inventory) PersonalityNames(ctx context.Context) ([]string, error) {
	ps := make([]*Personality, len(i.Personality))
	err := nds.GetMulti(ctx, i.Personality, ps)
	if err != nil {
		return nil, err
	}

	var pns []string
	for _, p := range ps {
		pns = append(pns, p.CanonicalName)
	}
	return pns, nil
}

func (i Inventory) ToString(ctx context.Context) (string, error) {
	pns, err := i.PersonalityNames(ctx)
	if err != nil {
		return "", nil
	}

	fileNameString := ""
	if tgapi.IsContributor(ctx) {
		fileNameString = i.FileName + "\n"
	}

	return fmt.Sprintf("%sUniqueID: %s\n%x (%d bytes)\n[%s]", fileNameString, i.FileUniqueID, i.MD5Sum, i.FileSize, strings.Join(pns, ", ")), nil
}

func (i Inventory) SendToChat(ctx context.Context, chatID int64) error {
	caption, err := i.ToString(ctx)
	if err != nil {
		return err
	}
	_, err = tgapi.BotFromContext(ctx).Send(tgapi.MakeFileable(chatID, i.FileID, i.FileType, caption))
	return err
}

const inventoryCurrentModelVersion = 1

func inventoryKey(ctx context.Context, fileUniqueID string) *datastore.Key {
	return datastore.NewKey(ctx, inventoryEntityKind, fileUniqueID, 0, nil)
}

func readWriteInventory(ctx context.Context, fileUniqueID string, modifier func(ctx context.Context, i *Inventory) error) (*Inventory, error) {
	i := new(Inventory)
	key := inventoryKey(ctx, fileUniqueID)

	err := nds.RunInTransaction(ctx, func(ctx context.Context) error {
		if err := nds.Get(ctx, key, i); err != nil {
			return err
		}

		if err := modifier(ctx, i); err != nil {
			return err
		}

		_, err := nds.Put(ctx, key, i)
		return err
	}, &datastore.TransactionOptions{})
	if err != nil {
		return nil, err
	}

	return i, nil
}

func GetInventory(ctx context.Context, fileUniqueID string) (*Inventory, error) {
	i := new(Inventory)
	key := inventoryKey(ctx, fileUniqueID)
	err := nds.Get(ctx, key, i)
	return i, err
}

// If not found, returns (nil, datastore.ErrNoSuchEntity)
func TryGetInventoryByTgFile(ctx context.Context, tgFile *tgapi.TGFile) (*Inventory, error) {
	i, err := GetInventory(ctx, tgFile.FileUniqueID)
	if err == nil {
		return i, nil
	}
	if err != datastore.ErrNoSuchEntity {
		return nil, err
	}

	return getInventoryByFile(ctx, tgFile.FileID, tgFile.FileSize)
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

func CreateOrUpdateInventory(ctx context.Context, tgFile *tgapi.TGFile, personality []*datastore.Key) (*Inventory, error) {
	i := new(Inventory)

	err := nds.RunInTransaction(ctx, func(ctx context.Context) error {
		key := inventoryKey(ctx, tgFile.FileUniqueID)
		err := nds.Get(ctx, key, i)

		// This is an existing Inventory, only admins or original creator can update it.
		if err == nil && !(tgapi.IsAdmin(ctx) || i.Creator == tgapi.UserFromContext(ctx).ID) {
			return ErrorOnlyAdminCanUpdateInventory
		}
		if err != nil && err != datastore.ErrNoSuchEntity {
			return err
		}

		i.FileID = tgFile.FileID
		i.FileUniqueID = tgFile.FileUniqueID

		// For existing Inventory that we know the name, don't override.
		if i.FileName == "" {
			i.FileName = tgFile.FileName
		}

		i.FileType = tgFile.FileType
		i.Personality = personality
		i.LastUsed = time.Now()

		if i.Creator == 0 {
			i.Creator = tgapi.UserFromContext(ctx).ID
		}

		i.ModelVersion = inventoryCurrentModelVersion

		shouldScheduleMetadataUpdate := err == datastore.ErrNoSuchEntity

		if _, err := nds.Put(ctx, key, i); err != nil {
			return err
		}

		if shouldScheduleMetadataUpdate {
			// New File. Schedule metadata update.
			if err := scheduler.ScheduleUpdateFileMetadata(ctx, []string{tgFile.FileUniqueID}, metadatamode.Default); err != nil {
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
		q = q.Order("-UsageCount").Order("-LastUsed")
	case sortmode.UsageCountAsc:
		q = q.Order("UsageCount").Order("LastUsed")
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

func bulkGetInventories(ctx context.Context, keys []*datastore.Key) ([]*Inventory, error) {
	g := make([]*Inventory, len(keys))
	err := nds.GetMulti(ctx, keys, g)
	if err != nil {
		if tryFlattenDatastoreNoSuchEntityMultiError(err) != datastore.ErrNoSuchEntity {
			return nil, err
		}
	}

	var is []*Inventory
	for _, i := range g {
		if i != nil {
			is = append(is, i)
		}
	}

	return is, nil
}

func QueryInventories(ctx context.Context, personalities []*datastore.Key, sortMode sortmode.SortMode, lastCursor, queryID string) ([]*Inventory, string, error) {
	keys, newCursor, err := queryInventoryKeys(ctx, personalities, sortMode, lastCursor, queryID)
	if err != nil {
		return nil, "", err
	}

	is, err := bulkGetInventories(ctx, keys)
	if err != nil {
		return nil, "", err
	}
	return is, newCursor, nil
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
	_, err := readWriteInventory(ctx, fileUniqueID, func(ctx context.Context, i *Inventory) error {
		i.UsageCount += 1
		i.LastUsed = time.Now()
		return nil
	})
	return err
}

func CountInventories(ctx context.Context, personality *datastore.Key, untweetedOnly bool) (int, error) {
	query := datastore.NewQuery(inventoryEntityKind).KeysOnly()
	if personality != nil {
		query = query.Filter("Personality = ", personality)
	}
	if untweetedOnly {
		query = query.Filter("LastTweetTime = ", time.Time{})
	}
	return query.Count(ctx)
}

func ReplaceFileID(ctx context.Context, oldFileUniqueID string, newFile *tgapi.TGFile) (*Inventory, error) {
	i := new(Inventory)

	err := nds.RunInTransaction(ctx, func(ctx context.Context) error {
		oldKey := inventoryKey(ctx, oldFileUniqueID)
		if err := nds.Get(ctx, oldKey, i); err != nil {
			return err
		}

		i.FileID = newFile.FileID
		i.FileUniqueID = newFile.FileUniqueID
		i.FileName = newFile.FileName
		i.FileType = newFile.FileType

		i.MD5Sum = nil
		i.FileSize = 0
		i.TwitterMediaID = ""

		if err := nds.Delete(ctx, oldKey); err != nil {
			return err
		}
		if _, err := nds.Put(ctx, inventoryKey(ctx, newFile.FileUniqueID), i); err != nil {
			return err
		}

		return scheduler.ScheduleUpdateFileMetadata(ctx, []string{newFile.FileUniqueID}, metadatamode.Default)
	}, &datastore.TransactionOptions{XG: true})

	return i, err
}

func OverrideFileName(ctx context.Context, fileUniqueID, fileName string) error {
	_, err := readWriteInventory(ctx, fileUniqueID, func(ctx context.Context, i *Inventory) error {
		i.FileName = fileName
		return nil
	})
	return err
}

func RotateReservoir(ctx context.Context) error {
	if reservoir.TryRotateReservoir(ctx, maxItems) {
		return nil
	}

	keys, err := datastore.NewQuery(inventoryEntityKind).KeysOnly().Filter("FileType =", tgapi.FileTypeMPEG4GIF).GetAll(ctx, nil)
	if err != nil {
		return err
	}
	return reservoir.RefillReservoir(ctx, keys)
}

func SetTwitterMediaID(ctx context.Context, fileUniqueID, twitterMediaID string) (*Inventory, error) {
	return readWriteInventory(ctx, fileUniqueID, func(ctx context.Context, i *Inventory) error {
		i.TwitterMediaID = twitterMediaID
		return nil
	})
}

func UpdateLastTweetTime(ctx context.Context, fileUniqueID string) error {
	_, err := readWriteInventory(ctx, fileUniqueID, func(ctx context.Context, i *Inventory) error {
		i.LastTweetTime = time.Now()
		return nil
	})
	return err
}

func RandomInventories(ctx context.Context, count int) ([]*Inventory, error) {
	keys, err := reservoir.ReadReservoir(ctx, count)
	if err != nil {
		return nil, err
	}

	return bulkGetInventories(ctx, keys)
}

func LastTweetedInventories(ctx context.Context, limit int) ([]*Inventory, error) {
	keys, err := datastore.NewQuery(inventoryEntityKind).KeysOnly().Order("-LastTweetTime").Limit(limit).GetAll(ctx, nil)
	if err != nil {
		return nil, err
	}

	for i, j := 0, len(keys)-1; i < j; i, j = i+1, j-1 {
		keys[i], keys[j] = keys[j], keys[i]
	}

	return bulkGetInventories(ctx, keys)
}
