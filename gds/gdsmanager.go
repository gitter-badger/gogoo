// Package gds provides the basic APIs to communicate with Google datastore
package gds

import (
	"fmt"
	"reflect"

	log "github.com/cihub/seelog"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/cloud"
	"google.golang.org/cloud/datastore"
)

var gdsError = fmt.Errorf("GDS Error")

// BuildGdsContext builds the singlton context for GdsManager
func BuildGdsContext(serviceEmail string, key []byte, projectId string) (context.Context, *datastore.Client, error) {
	conf := &jwt.Config{
		Email:      serviceEmail,
		PrivateKey: key,
		Scopes: []string{
			datastore.ScopeDatastore,
			datastore.ScopeUserEmail,
		},
		TokenURL: google.JWTTokenURL,
	}

	ctx := context.Background()
	client, err := datastore.NewClient(ctx, projectId, cloud.WithTokenSource(conf.TokenSource(ctx)))
	if err != nil {
		return ctx, nil, err
	}

	return ctx, client, nil
}

// GdsManager is for low level communication with Google datastore
type GdsManager struct {
	SuffixOfKind string

	Client *datastore.Client `inject:""`
}

func (manager *GdsManager) Setup(suffixOfKind string) {
	manager.SuffixOfKind = suffixOfKind
}

func (manager *GdsManager) BuildKey(kind, keyName string) *datastore.Key {
	return datastore.NewKey(context.Background(), kind, keyName, 0, nil)
}

// Put inserts/updates the entity
func (manager *GdsManager) Put(key *datastore.Key, entity interface{}) (*datastore.Key, error) {
	log.Tracef("Put entity: key[%s]", key.Name())

	var resultKey *datastore.Key
	if key, err := manager.Client.Put(context.Background(), key, entity); err != nil {
		return nil, err
	} else {
		resultKey = key
	}

	// Use reflection to setup key of entity
	f := reflect.ValueOf(entity).Elem().FieldByName("Key")
	if f.IsValid() && f.CanSet() {
		f.Set(reflect.ValueOf(key))
	}

	return resultKey, nil
}

// PutUnique inserts/updates entity with unique key (if the same key existed, issue error)
func (manager *GdsManager) PutUnique(key *datastore.Key, entity interface{}) error {
	log.Tracef("PutUnique entity: key[%s]", key.Name())

	tx := manager.GetTx()

	if err := tx.Get(key, entity); err == nil {
		return fmt.Errorf("Unique condition violation")
	}

	_, err := tx.Put(key, entity)
	if err != nil {
		return err
	}

	if _, err := tx.Commit(); err != nil {
		log.Warnf("%s", err.Error())
		return err
	}

	return nil
}

// Get gets the entity by key
func (manager *GdsManager) Get(key *datastore.Key, entity interface{}) error {
	log.Tracef("Get entity: key[%s]", key.Name())

	err := manager.Client.Get(context.Background(), key, entity)
	if err != nil {
		log.Tracef("Error: %s: kind[%s], key[%s]", err.Error(), key.Kind(), key.Name())

		return gdsError
	}

	// Use reflection to setup key of entity
	f := reflect.ValueOf(entity).Elem().FieldByName("Key")
	if f.IsValid() && f.CanSet() {
		f.Set(reflect.ValueOf(key))
	}

	return nil
}

// GetMulti gets the entities by keys
func (manager *GdsManager) GetMulti(keys []*datastore.Key, dst interface{}) error {
	err := manager.Client.GetMulti(context.Background(), keys, dst)
	if err != nil {
		log.Tracef("Error: %s", err.Error())

		return gdsError
	}

	return nil
}

// GetKeysOnly gets only keys bu query
func (manager *GdsManager) GetKeysOnly(query *datastore.Query) ([]*datastore.Key, error) {
	query = query.KeysOnly()

	type Any struct{}
	result := &[]Any{}
	if keys, err := manager.GetAll(query, result); err != nil {
		return nil, err
	} else {
		return keys, nil
	}
}

// Delete deletes the entity by key (if the entity is not existed, there is no error)
func (manager *GdsManager) Delete(key *datastore.Key) error {
	if key == nil {
		return fmt.Errorf("key is nil")
	}

	log.Tracef("Delete entity: key[%s]", key.Name())

	err := manager.Client.Delete(context.Background(), key)
	if err != nil {
		log.Tracef("Error: %s", err.Error())

		return gdsError
	}

	return nil
}

// GetAll fetchs all entities by the query. The parameter `result` should be type of `*[]*<Entity>`
func (manager *GdsManager) GetAll(query *datastore.Query, result interface{}) ([]*datastore.Key, error) {
	log.Trace("Get all by query")

	keys, err := manager.Client.GetAll(context.Background(), query, result)
	if err != nil {
		log.Warnf("Error: %s", err.Error())

		return nil, gdsError
	}

	// Use reflection to setup keys of entities
	s := reflect.ValueOf(result).Elem()
	for i := 0; i < s.Len(); i++ {
		f := s.Index(i).Elem().FieldByName("Key")
		if f.IsValid() && f.CanSet() {
			f.Set(reflect.ValueOf(keys[i]))
		}
	}

	return keys, nil
}

// GetCount return count of result
func (manager *GdsManager) GetCount(query *datastore.Query) (int, error) {
	log.Trace("Get count by query")

	count, err := manager.Client.Count(context.Background(), query)
	if err != nil {
		log.Warnf("Error: %s", err.Error())

		return 0, gdsError
	}

	return count, nil
}

// DeleteAll deletes all entities under some Kind
func (manager *GdsManager) DeleteAll(kindName string) error {
	log.Trace("Delete all")

	query := datastore.NewQuery(kindName).KeysOnly()

	type Any struct{}
	result := &[]Any{}

	keys, err := manager.GetAll(query, result)
	if err != nil {
		return gdsError
	}

	for _, key := range keys {
		err = manager.Delete(key)
		if err != nil {
			return gdsError
		}
	}

	return nil
}

// GetTx gets the datastore transaction
func (manager *GdsManager) GetTx() *datastore.Transaction {
	tx, _ := manager.Client.NewTransaction(context.Background(), datastore.Serializable)
	return tx
}
