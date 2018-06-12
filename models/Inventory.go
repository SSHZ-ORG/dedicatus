package models

import (
	"fmt"
	"strings"

	"github.com/SSHZ-ORG/dedicatus/utils"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

type Inventory struct {
	FileID      string
	FileType    string
	Personality []*datastore.Key
	UsageCount  int64
}

func (i Inventory) ToString(ctx context.Context) (string, error) {
	ps := make([]*Personality, len(i.Personality))
	err := datastore.GetMulti(ctx, i.Personality, ps)
	if err != nil {
		return "", err
	}

	var pns []string
	for _, p := range ps {
		pns = append(pns, p.CanonicalName)
	}

	return fmt.Sprintf("%s [%s]", i.FileID, strings.Join(pns, ", ")), nil
}

func GetInventory(ctx context.Context, fileID string) (*Inventory, error) {
	i := new(Inventory)
	key := datastore.NewKey(ctx, inventoryEntityKind, fileID, 0, nil)
	err := datastore.Get(ctx, key, i)
	return i, err
}

func CreateInventory(ctx context.Context, fileID string, personality []*datastore.Key) (*Inventory, error) {
	key := datastore.NewKey(ctx, inventoryEntityKind, fileID, 0, nil)

	i := &Inventory{
		FileID:      fileID,
		FileType:    utils.FileTypeMPEG4GIF,
		Personality: personality,
		UsageCount:  0,
	}

	_, err := datastore.Put(ctx, key, i)
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
