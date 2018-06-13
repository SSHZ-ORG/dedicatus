package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/SSHZ-ORG/dedicatus/utils"
	"github.com/qedus/nds"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

type Inventory struct {
	FileID      string
	FileType    string
	Personality []*datastore.Key
	Creator     int

	UsageCount int64
	LastUsed   time.Time
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

	return fmt.Sprintf("%s [%s]", i.FileID, strings.Join(pns, ", ")), nil
}

func inventoryKey(ctx context.Context, fileID string) *datastore.Key {
	return datastore.NewKey(ctx, inventoryEntityKind, fileID, 0, nil)
}

func GetInventory(ctx context.Context, fileID string) (*Inventory, error) {
	i := new(Inventory)
	key := inventoryKey(ctx, fileID)
	err := nds.Get(ctx, key, i)
	return i, err
}

func CreateInventory(ctx context.Context, fileID string, personality []*datastore.Key, userID int) (*Inventory, error) {
	i := new(Inventory)
	key := inventoryKey(ctx, fileID)
	nds.Get(ctx, key, i)

	i.FileID = fileID
	i.FileType = utils.FileTypeMPEG4GIF
	i.Personality = personality
	i.Creator = userID
	i.LastUsed = time.Now()

	_, err := nds.Put(ctx, key, i)
	return i, err
}

func FindInventories(ctx context.Context, personality *datastore.Key, lastCursor string) ([]*Inventory, string) {
	var inventories []*Inventory

	q := datastore.NewQuery(inventoryEntityKind).Filter("Personality = ", personality).Order("-UsageCount")

	if (lastCursor != "") {
		cursor, err := datastore.DecodeCursor(string(lastCursor))
		if err == nil {
			q = q.Start(cursor)
		}
	}

	t := q.Limit(50).Run(ctx)
	for {
		var i Inventory
		_, err := t.Next(&i)
		if err == datastore.Done {
			break
		}
		if err != nil {
			log.Errorf(ctx, "fetching next Inventory: %v", err)
			break
		}
		inventories = append(inventories, &i)
	}

	nextCursor := ""
	if cursor, err := t.Cursor(); err == nil {
		nextCursor = cursor.String()
	}

	return inventories, nextCursor
}

func GloballyLastUsedInventories(ctx context.Context) ([]*Inventory, error) {
	var inventories []*Inventory
	_, err := datastore.NewQuery(inventoryEntityKind).Order("-LastUsed").Limit(20).GetAll(ctx, &inventories)
	return inventories, err
}

func IncrementUsageCounter(ctx context.Context, fileID string) error {
	i := new(Inventory)
	key := inventoryKey(ctx, fileID)
	err := nds.Get(ctx, key, i)
	if err != nil {
		return err
	}

	i.UsageCount += 1
	i.LastUsed = time.Now()

	_, err = nds.Put(ctx, key, i)
	return err
}
