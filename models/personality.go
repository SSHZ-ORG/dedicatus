package models

import (
	"context"
	"fmt"
	"strings"

	"github.com/SSHZ-ORG/dedicatus/kgapi"
	"github.com/qedus/nds"
	"google.golang.org/appengine/v2/datastore"
)

type Personality struct {
	KGID          string
	CanonicalName string

	TwitterUserID string
}

func (p Personality) ToString(ctx context.Context, key *datastore.Key) (string, error) {
	as, err := findAliasForPersonality(ctx, key)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("kg:%s %s as (%s)", p.KGID, p.CanonicalName, strings.Join(as, ", ")), nil
}

// Stored in TwitterUserID field for Personalities that do not have a Twitter account.
const PersonalityNoTwitterAccountPlaceholder = "-"

func CreatePersonality(ctx context.Context, KGID string, name string) (*datastore.Key, *Personality, error) {
	p := &Personality{
		KGID:          KGID,
		CanonicalName: name,
	}

	var key *datastore.Key

	err := nds.RunInTransaction(ctx, func(ctx context.Context) error {
		k, err := nds.Put(ctx, datastore.NewIncompleteKey(ctx, personalityEntityKind, nil), p)
		if err != nil {
			return err
		}

		_, err = AddAlias(ctx, p.CanonicalName, k)
		if err != nil {
			return err
		}

		key = k
		return nil
	}, &datastore.TransactionOptions{XG: true})

	if err != nil {
		return nil, nil, err
	}

	return key, p, nil
}

func GetPersonalities(ctx context.Context, keys []*datastore.Key) ([]*Personality, error) {
	ps := make([]*Personality, len(keys))
	err := nds.GetMulti(ctx, keys, ps)
	return ps, err
}

func GetPersonalityByName(ctx context.Context, name string) (*datastore.Key, *Personality, error) {
	var ps []*Personality
	keys, err := datastore.NewQuery(personalityEntityKind).Filter("CanonicalName = ", name).Limit(1).GetAll(ctx, &ps)
	if len(keys) == 1 {
		return keys[0], ps[0], nil
	}
	return nil, nil, err
}

func GetPersonalityByKGID(ctx context.Context, KGID string) (*datastore.Key, *Personality, error) {
	var ps []*Personality
	keys, err := datastore.NewQuery(personalityEntityKind).Filter("KGID = ", KGID).Limit(1).GetAll(ctx, &ps)
	if len(keys) == 1 {
		return keys[0], ps[0], nil
	}
	return nil, nil, err
}

func TryFindOnlyPersonality(ctx context.Context, query string) (*datastore.Key, error) {
	keys, err := TryFindPersonalitiesByAlias(ctx, query)
	if err != nil {
		return nil, err
	}
	if len(keys) == 1 {
		return keys[0], nil
	}
	return nil, nil
}

func TryFindPersonalitiesWithKG(ctx context.Context, query string) ([]*datastore.Key, error) {
	// Do we know this personality?
	keys, err := TryFindPersonalitiesByAlias(ctx, query)
	if err != nil {
		return nil, err
	}
	if len(keys) > 0 {
		return keys, nil
	}

	// Let's ask Google
	if KGID := kgapi.TryFindKGEntity(ctx, query); KGID != "" {
		keys, err := datastore.NewQuery(personalityEntityKind).Filter("KGID = ", KGID).Limit(1).KeysOnly().GetAll(ctx, nil)
		if len(keys) == 1 {
			return keys, nil
		}
		return nil, err
	}

	return nil, nil
}

func ListAllPersonalities(ctx context.Context) ([]*datastore.Key, error) {
	return datastore.NewQuery(personalityEntityKind).KeysOnly().GetAll(ctx, nil)
}

func UpdateTwitterUserID(ctx context.Context, key *datastore.Key, userID string) error {
	return nds.RunInTransaction(ctx, func(ctx context.Context) error {
		p := new(Personality)
		if err := nds.Get(ctx, key, p); err != nil {
			return err
		}

		p.TwitterUserID = userID

		_, err := nds.Put(ctx, key, p)
		return err
	}, &datastore.TransactionOptions{})
}
