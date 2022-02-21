package models

import (
	"context"
	"fmt"
	"strings"

	"github.com/qedus/nds"
	"google.golang.org/appengine/v2/datastore"
)

const tagCurrentModelVersion = 0

type Tag struct {
	// Name of the tag. The first one is canonical.
	Name []string

	ModelVersion int
}

func (t *Tag) CanonicalName() string {
	if t == nil {
		return "(nil Tag)"
	}
	return "#" + t.Name[0]
}

func (t *Tag) ToString() string {
	if t == nil {
		return "(nil Tag)"
	}
	return fmt.Sprintf("#[%s]", strings.Join(t.Name, ", "))
}

// If not found, returns nil, nil, nil
func FindTag(ctx context.Context, name string) (*datastore.Key, *Tag, error) {
	var ts []*Tag
	keys, err := datastore.NewQuery(tagEntityKind).Filter("Name =", name).Limit(1).GetAll(ctx, &ts)
	if len(ts) == 1 {
		return keys[0], ts[0], nil
	}
	return nil, nil, err
}

// If not found, returns nil, nil, ErrNoSuchEntity
func mustFindTag(ctx context.Context, name string) (*datastore.Key, *Tag, error) {
	k, t, err := FindTag(ctx, name)
	if err != nil {
		return nil, nil, err
	}
	if k == nil {
		return nil, nil, datastore.ErrNoSuchEntity
	}
	return k, t, nil
}

// If the tag with the same name already exists, it will be returned.
func CreateTag(ctx context.Context, name string) (*datastore.Key, *Tag, error) {
	k, t, err := FindTag(ctx, name)
	if err != nil {
		return nil, nil, err
	}

	if t != nil {
		return k, t, nil
	}

	t = &Tag{
		Name:         []string{name},
		ModelVersion: tagCurrentModelVersion,
	}
	k, err = nds.Put(ctx, datastore.NewIncompleteKey(ctx, tagEntityKind, nil), t)
	return k, t, err
}

func AddTagName(ctx context.Context, name, newName string) (*datastore.Key, *Tag, error) {
	k, t, err := mustFindTag(ctx, name)
	if err != nil {
		return nil, nil, err
	}

	for _, n := range t.Name {
		if n == newName {
			return k, t, nil
		}
	}
	t.Name = append(t.Name, newName)
	_, err = nds.Put(ctx, k, t)
	return k, t, err
}

func DeleteTag(ctx context.Context, name string) (*datastore.Key, *Tag, error) {
	k, t, err := mustFindTag(ctx, name)
	if err != nil {
		return nil, nil, err
	}

	for i, n := range t.Name {
		if n == name {
			t.Name = append(t.Name[:i], t.Name[i+1:]...)
			break
		}
	}

	if len(t.Name) > 0 {
		_, err = nds.Put(ctx, k, t)
	} else {
		err = nds.Delete(ctx, k)
		k, t = nil, nil
	}
	return k, t, err
}

func SetTagCanonicalName(ctx context.Context, name string) (*datastore.Key, *Tag, error) {
	k, t, err := mustFindTag(ctx, name)
	if err != nil {
		return nil, nil, err
	}

	for i, n := range t.Name {
		if n == name {
			t.Name = append(t.Name[:i], t.Name[i+1:]...)
			break
		}
	}

	t.Name = append([]string{name}, t.Name...)
	_, err = nds.Put(ctx, k, t)
	return k, t, err
}
