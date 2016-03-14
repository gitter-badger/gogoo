// Package pubsub communicates with pub/sub.
// https://godoc.org/google.golang.org/cloud/pubsub.
// https://godoc.org/google.golang.org/api/pubsub/v1.
package pubsub

import (
	"log"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	pbsb "google.golang.org/api/pubsub/v1"
)

// BuildPbsbService builds the singlton service for PbsbManager
func BuildPbsbService(serviceEmail string, key []byte) (*pbsb.Service, error) {
	conf := &jwt.Config{
		Email:      serviceEmail,
		PrivateKey: key,
		Scopes: []string{
			pbsb.CloudPlatformScope,
			pbsb.PubsubScope,
		},
		TokenURL: google.JWTTokenURL,
	}

	if service, err := pbsb.New(conf.Client(oauth2.NoContext)); err != nil {
		return nil, err
	} else {
		return service, nil
	}
}

// PbsbManager communicates with google cloud pub/sub
type PbsbManager struct {
	Service       *pbsb.Service `inject:""`
	topicsService *pbsb.ProjectsTopicsService
}

func (manager *PbsbManager) Setup() {
	manager.topicsService = pbsb.NewProjectsTopicsService(manager.Service)
}

func (manager *PbsbManager) ListTopics(projectId string) {
	response, err := manager.topicsService.List(projectId).Do()
	if err != nil {
		log.Printf("err: %+v", err)
		return
	}

	log.Printf("topics: %+v", response.Topics[0].Name)
}
