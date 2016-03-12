// Package replicapoolupdater provides APIs to communicate with Google autoscaling service
package replicapoolupdater

import (
	"fmt"

	log "github.com/cihub/seelog"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"

	rpu "google.golang.org/api/replicapoolupdater/v1beta1"
)

var rollingUpdateError = fmt.Errorf("RollingUpdate Error")

// BuildRpuService builds the singlton service for RpuService
func BuildRpuService(serviceEmail string, key []byte) (*rpu.Service, error) {
	conf := &jwt.Config{
		Email:      serviceEmail,
		PrivateKey: key,
		Scopes: []string{
			rpu.ReplicapoolScope,
		},
		TokenURL: google.JWTTokenURL,
	}

	if service, err := rpu.New(conf.Client(oauth2.NoContext)); err != nil {
		return nil, err
	} else {
		return service, nil
	}
}

// RpuManager
//
// https://godoc.org/google.golang.org/api/replicapoolupdater/v1beta1
type RpuManager struct {
	Service *rpu.Service `inject:""`
}

// Insert starts rolling update for instances in instance group
//
// https://godoc.org/google.golang.org/api/replicapoolupdater/v1beta1#RollingUpdatesService.Insert
func (manager *RpuManager) Insert(projectId, zone string, rollingUpdate *rpu.RollingUpdate) (*rpu.Operation, error) {
	log.Tracef("Insert: projectId[%s], zone[%s]", projectId, zone)

	rollingUpdateService := rpu.NewRollingUpdatesService(manager.Service)

	op, err := rollingUpdateService.Insert(projectId, zone, rollingUpdate).Do()
	if err != nil {
		log.Warnf("Error: %s", err.Error())

		return nil, rollingUpdateError
	}

	return op, nil
}

// List lists recent rolling updates
//
// https://godoc.org/google.golang.org/api/replicapoolupdater/v1beta1#RollingUpdatesService.List
func (manager *RpuManager) List(projectId, zone string) (*rpu.RollingUpdateList, error) {
	log.Tracef("List: projectId[%s], zone[%s]", projectId, zone)

	rollingUpdateService := rpu.NewRollingUpdatesService(manager.Service)

	list, err := rollingUpdateService.List(projectId, zone).Do()
	if err != nil {
		log.Warnf("Error: %s", err.Error())

		return nil, rollingUpdateError
	}

	return list, nil
}

// Rollback rollbacks specified rolling update
//
// https://godoc.org/google.golang.org/api/replicapoolupdater/v1beta1#RollingUpdatesService.Rollback
func (manager *RpuManager) Rollback(projectId, zone, rollingUpdateId string) (*rpu.Operation, error) {
	log.Tracef("Rollback: projectId[%s], zone[%s], rollingUpdateId[%s]", projectId, zone, rollingUpdateId)

	rollingUpdateService := rpu.NewRollingUpdatesService(manager.Service)

	op, err := rollingUpdateService.Rollback(projectId, zone, rollingUpdateId).Do()
	if err != nil {
		log.Warnf("Error: %s", err.Error())

		return nil, rollingUpdateError
	}

	return op, nil
}
