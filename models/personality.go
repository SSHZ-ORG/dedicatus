package models

import (
	"fmt"
	"strings"

	"github.com/SSHZ-ORG/dedicatus/utils"
	"github.com/qedus/nds"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/memcache"
)

const (
	findByNicknameMemcacheKey = "FPN1:"
)

func getFindByNicknameMemcacheKey(query string) string {
	return findByNicknameMemcacheKey + query
}

func getFindByNicknameMemcache(ctx context.Context, query string) (*datastore.Key, error) {
	item, err := memcache.Get(ctx, getFindByNicknameMemcacheKey(query))
	if err != nil {
		return nil, err
	}
	if len(item.Value) == 0 {
		return nil, nil
	}
	return datastore.DecodeKey(string(item.Value))
}

func setFindByNicknameMemcache(ctx context.Context, query string, result *datastore.Key) {
	value := []byte{}
	if result != nil {
		value = []byte(result.Encode())
	}
	memcache.Set(ctx, &memcache.Item{
		Key:   getFindByNicknameMemcacheKey(query),
		Value: value,
	})
}

func DeleteFindByNicknameMemcache(ctx context.Context, query string) {
	memcache.Delete(ctx, getFindByNicknameMemcacheKey(query))
}

type Personality struct {
	KGID          string
	CanonicalName string
	Nickname      []string
}

func (p Personality) ToString() string {
	return fmt.Sprintf("kg:%s %s (%s)", p.KGID, p.CanonicalName, strings.Join(p.Nickname, ", "))
}

func CreatePersonality(ctx context.Context, KGID string, name string) (*Personality, error) {
	key := datastore.NewIncompleteKey(ctx, personalityEntityKind, nil)

	p := &Personality{
		KGID:          KGID,
		CanonicalName: name,
		Nickname:      []string{name},
	}

	_, err := nds.Put(ctx, key, p)
	return p, err
}

func GetPersonality(ctx context.Context, key *datastore.Key) (*Personality, error) {
	p := new(Personality)
	err := nds.Get(ctx, key, p)
	return p, err
}

func GetPersonalityByName(ctx context.Context, name string) (*datastore.Key, *Personality, error) {
	var ps []*Personality
	keys, err := datastore.NewQuery(personalityEntityKind).Filter("CanonicalName = ", name).Limit(1).GetAll(ctx, &ps)
	if len(keys) == 1 {
		return keys[0], ps[0], nil
	}
	return nil, nil, err
}

func GetPersonalityByKGID(ctx context.Context, KGID string) (*Personality, error) {
	var ps []*Personality
	keys, err := datastore.NewQuery(personalityEntityKind).Filter("KGID = ", KGID).Limit(1).GetAll(ctx, &ps)
	if len(keys) == 1 {
		return ps[0], nil
	}
	return nil, err
}

func TryFindPersonality(ctx context.Context, query string) (*datastore.Key, error) {
	if resultFromMemcache, err := getFindByNicknameMemcache(ctx, query); err == nil {
		return resultFromMemcache, err
	}

	keys, err := datastore.NewQuery(personalityEntityKind).Filter("Nickname = ", strings.ToLower(query)).Limit(1).KeysOnly().GetAll(ctx, nil)
	if err != nil {
		return nil, err
	}

	var key *datastore.Key
	if len(keys) == 1 {
		key = keys[0]
	}
	setFindByNicknameMemcache(ctx, query, key)
	return key, nil
}

func TryFindPersonalityWithKG(ctx context.Context, query string) (*datastore.Key, error) {
	// Do we know this personality?
	key, err := TryFindPersonality(ctx, query)
	if err != nil {
		return nil, err
	}
	if key != nil {
		return key, nil
	}

	// Let's ask Google
	KGID, err := utils.TryFindKGEntity(ctx, query)
	if err != nil {
		return nil, err
	}
	if KGID != "" {
		keys, err := datastore.NewQuery(personalityEntityKind).Filter("KGID = ", KGID).Limit(1).KeysOnly().GetAll(ctx, nil)
		if len(keys) == 1 {
			return keys[0], nil
		}
		return nil, err
	}

	return nil, nil
}
