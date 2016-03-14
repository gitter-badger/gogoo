package cloudsql_test

import (
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/browny/gogoo/cloudsql"
	"github.com/browny/gogoo/config"

	"github.com/facebookgo/inject"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	sql "google.golang.org/api/sqladmin/v1beta4"
)

var testedCloudSqlManager cloudsql.CloudSqlManager
var testedProjectId string
var testedZone string

func TestCloudSqlManagerTestSuite(t *testing.T) {
	suite.Run(t, new(CloudSqlManagerTestSuite))
}

type CloudSqlManagerTestSuite struct {
	suite.Suite
}

func (suite *CloudSqlManagerTestSuite) SetupSuite() {
	gcloudConfig := config.LoadGcloudConfig(config.LoadAsset("/config/config.json"))
	key, _ := ioutil.ReadAll(config.LoadAsset("/config/key.pem"))

	// Construct dependency graph
	sqlService, _ := cloudsql.BuildCloudSqlService(
		gcloudConfig.ServiceAccount, key)

	var g inject.Graph
	err := g.Provide(
		&inject.Object{Value: sqlService},
		&inject.Object{Value: &testedCloudSqlManager},
	)
	if err != nil {
		os.Exit(1)
	}
	if err := g.Populate(); err != nil {
		os.Exit(1)
	}
	// :~)

	testedProjectId = gcloudConfig.ProjectId
	testedZone = "asia-east1-b"

	log.Println("======== SetupSuite  ========")
}

func (suite *CloudSqlManagerTestSuite) Test01_GetDatabase() {
	dbInstance, _ := testedCloudSqlManager.GetDatabase(testedProjectId, "frontend-staging")
	assert.Equal(suite.T(), "frontend-staging", dbInstance.Name)
}

func (suite *CloudSqlManagerTestSuite) Test02_PatchAclEntriesOfDatabase() {
	entries := []*sql.AclEntry{
		&sql.AclEntry{
			Kind:  "sql#aclEntry",
			Name:  "test",
			Value: "1.1.1.2",
		}}

	if _, err := testedCloudSqlManager.PatchAclEntriesOfDatabase(testedProjectId, "test-database", entries); err != nil {
		log.Printf("err: %s", err.Error())
	}
}

func (suite *CloudSqlManagerTestSuite) Test03_GetFilteredAclEntriesOfDatabase() {
	isContain := func(checked string) bool {
		return strings.Contains(checked, "frontend")
	}

	aclEntries, _ := testedCloudSqlManager.GetFilteredAclEntriesOfDatabase(testedProjectId, "frontend", isContain)
	for _, entry := range aclEntries {
		log.Printf("entry: name[%s]", entry.Name)
	}
}
