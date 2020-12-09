package models

import (
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

const (
	personalityEntityKind = "Personality"
	inventoryEntityKind   = "Inventory"
	aliasEntityKind       = "Alias"
	tagEntityKind         = "Tag"
)

// If err is a MultiError, and everything inside is ErrNoSuchEntity, return as ErrNoSuchEntity
func tryFlattenDatastoreNoSuchEntityMultiError(err error) error {
	// Is this a MultiError?
	if multiErrors, ok := err.(appengine.MultiError); ok {
		for _, e := range multiErrors {
			if e != nil && e != datastore.ErrNoSuchEntity {
				return err
			}
		}
		return datastore.ErrNoSuchEntity
	}
	// Not a MultiError, return back
	return err
}
