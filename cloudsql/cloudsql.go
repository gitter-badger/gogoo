// Package cloudsql provides the basic APIs to communicate with Google Cloud SQL
package cloudsql

import (
	"fmt"

	log "github.com/cihub/seelog"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	sql "google.golang.org/api/sqladmin/v1beta4"
)

// BuildCloudSqlService builds the singlton service for CloudSql
func BuildCloudSqlService(serviceEmail string, key []byte) (*sql.Service, error) {
	conf := &jwt.Config{
		Email:      serviceEmail,
		PrivateKey: key,
		Scopes: []string{
			sql.SqlserviceAdminScope,
		},
		TokenURL: google.JWTTokenURL,
	}

	if service, err := sql.New(conf.Client(oauth2.NoContext)); err != nil {
		return nil, err
	} else {
		return service, nil
	}
}

// CloudSqlManager
//
// https://godoc.org/google.golang.org/api/sqladmin/v1beta4
type CloudSqlManager struct {
	Service *sql.Service `inject:""`
}

// GetDatabase gets the database instance
func (manager *CloudSqlManager) GetDatabase(projectId, dbName string) (*sql.DatabaseInstance, error) {
	log.Tracef("GetDatabase: projectId[%s], db[%s]", projectId, dbName)

	dbInstanceService := sql.NewInstancesService(manager.Service)
	if dbInstanceService == nil {
		return nil, fmt.Errorf("Fail NewInstancesService")
	}

	if dbInstance, err := dbInstanceService.Get(projectId, dbName).Do(); err != nil {
		return nil, err
	} else {
		for _, an := range dbInstance.Settings.IpConfiguration.AuthorizedNetworks {
			log.Debugf("dbInstance: %+v", an)
		}
		return dbInstance, nil
	}
}

// PatchAclEntriesOfDatabase updates the aclEntries settings of the database instance
func (manager *CloudSqlManager) PatchAclEntriesOfDatabase(projectId, dbName string, entries []*sql.AclEntry) (*sql.Operation, error) {
	log.Debugf("PatchAclEntriesOfDatabase: projectId[%s], db[%s]", projectId, dbName)

	dbInstance, err := manager.GetDatabase(projectId, dbName)
	if err != nil {
		return nil, err
	}

	dbInstance.Settings.IpConfiguration.AuthorizedNetworks = entries

	dbInstanceService := sql.NewInstancesService(manager.Service)
	if dbInstanceService == nil {
		return nil, fmt.Errorf("Fail NewInstancesService")
	}

	return dbInstanceService.Patch(projectId, dbName, dbInstance).Do()
}

// GetFilteredAclEntriesOfDatabase gets aclEntries which satisfies entry name filter
func (manager *CloudSqlManager) GetFilteredAclEntriesOfDatabase(
	projectId, dbName string, notContain func(string) bool) ([]*sql.AclEntry, error) {

	log.Debugf("GetFilteredAclEntriesOfDatabase: projectId[%s], db[%s]", projectId, dbName)

	dbInstance, err := manager.GetDatabase(projectId, dbName)
	if err != nil {
		return nil, err
	}

	aclEntries := []*sql.AclEntry{}
	entries := dbInstance.Settings.IpConfiguration.AuthorizedNetworks
	for _, entry := range entries {
		if notContain(entry.Name) {
			aclEntries = append(aclEntries, entry)
		}
	}

	return aclEntries, nil
}
