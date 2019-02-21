package models

import (
	"fmt"
	"strings"

	"github.com/SSHZ-ORG/dedicatus/utils"
	"github.com/qedus/nds"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

type Alias struct {
	Name        string
	Personality []*datastore.Key
}

func (a Alias) ToString(ctx context.Context) (string, error) {
	ps := make([]*Personality, len(a.Personality))
	err := nds.GetMulti(ctx, a.Personality, ps)
	if err != nil {
		return "", err
	}

	var pns []string
	for _, p := range ps {
		pns = append(pns, p.CanonicalName)
	}

	return fmt.Sprintf("%s [%s]", a.Name, strings.Join(pns, ", ")), nil
}

func getAliasKey(ctx context.Context, name string) *datastore.Key {
	return datastore.NewKey(ctx, aliasEntityKind, name, 0, nil)
}

func TryFindPersonalitiesByAlias(ctx context.Context, query string) ([]*datastore.Key, error) {
	query = strings.ToLower(query)

	a := new(Alias)
	err := nds.Get(ctx, getAliasKey(ctx, query), a)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			return nil, nil
		}
		return nil, err
	}
	return a.Personality, nil
}

func AddAlias(ctx context.Context, alias string, personality *datastore.Key) (*Alias, error) {
	alias = strings.ToLower(alias)

	a := new(Alias)
	k := getAliasKey(ctx, alias)
	err := nds.Get(ctx, k, a)
	if err != nil && err != datastore.ErrNoSuchEntity {
		return nil, err
	}

	a.Name = alias
	if utils.KeyContains(a.Personality, personality) == -1 {
		a.Personality = append(a.Personality, personality)
	}

	_, err = nds.Put(ctx, k, a)
	return a, err
}

func DeleteAlias(ctx context.Context, alias string, personality *datastore.Key) (*Alias, error) {
	alias = strings.ToLower(alias)

	a := new(Alias)
	k := getAliasKey(ctx, alias)
	err := nds.Get(ctx, k, a)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			return nil, nil
		}
		return nil, err
	}

	idx := utils.KeyContains(a.Personality, personality)
	if idx != -1 {
		a.Personality[idx] = a.Personality[len(a.Personality)-1]
		a.Personality[len(a.Personality)-1] = nil
		a.Personality = a.Personality[:len(a.Personality)-1]
	}

	if len(a.Personality) > 0 {
		_, err = nds.Put(ctx, k, a)
		return a, err
	} else {
		return nil, nds.Delete(ctx, k)
	}
}

func findAliasForPersonality(ctx context.Context, personality *datastore.Key) ([]string, error) {
	keys, err := datastore.NewQuery(aliasEntityKind).KeysOnly().Filter("Personality = ", personality).GetAll(ctx, nil)
	if err != nil {
		return nil, err
	}

	alias := make([]*Alias, len(keys))
	err = nds.GetMulti(ctx, keys, alias)
	if err != nil {
		return nil, err
	}

	var as []string
	for _, a := range alias {
		as = append(as, a.Name)
	}
	return as, nil
}
