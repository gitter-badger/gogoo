// Package storage communicates with google storage
package storage

import (
	"log"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	storage "google.golang.org/api/storage/v1"
)

// BuildStorageService builds the singlton service for Storage
func BuildStorageService(serviceEmail string, key []byte) (*storage.Service, error) {
	conf := &jwt.Config{
		Email:      serviceEmail,
		PrivateKey: key,
		Scopes: []string{
			storage.DevstorageFullControlScope,
			storage.DevstorageReadWriteScope,
		},
		TokenURL: google.JWTTokenURL,
	}

	service, err := storage.New(conf.Client(oauth2.NoContext))
	if err != nil {
		return nil, err
	}

	return service, nil
}

type StorageManager struct {
	*storage.Service `inject:""`
	bucketService    *storage.BucketsService
}

func (manager *StorageManager) Setup() {
	manager.bucketService = storage.NewBucketsService(manager.Service)
}

func (manager *StorageManager) ListBuckets(projectID string) {
	buckets, err := manager.bucketService.List(projectID).Do()
	if err != nil {
		log.Printf("err: %+v", err)
		return
	}

	log.Printf("buckets: %+v", buckets.Items[0].Name)
}
