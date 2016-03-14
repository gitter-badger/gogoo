package storage_test

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"gogoo/config"
	"gogoo/storage"

	"github.com/facebookgo/inject"
	"github.com/stretchr/testify/suite"
)

var testedStorageManager storage.StorageManager
var testedProjectId string
var testedZone string

func TestStorageManagerTestSuite(t *testing.T) {
	suite.Run(t, new(StorageManagerTestSuite))
}

type StorageManagerTestSuite struct {
	suite.Suite
}

func (suite *StorageManagerTestSuite) SetupSuite() {
	gcloudConfig := config.LoadGcloudConfig(config.LoadAsset("/config/config.json"))
	key, _ := ioutil.ReadAll(config.LoadAsset("/config/key.pem"))

	// Construct dependency graph
	storageService, _ := storage.BuildStorageService(gcloudConfig.ServiceAccount, key)

	var g inject.Graph
	err := g.Provide(
		&inject.Object{Value: storageService},
		&inject.Object{Value: &testedStorageManager},
	)
	if err != nil {
		os.Exit(1)
	}
	if err := g.Populate(); err != nil {
		os.Exit(1)
	}
	// :~)

	testedStorageManager.Setup()

	testedProjectId = gcloudConfig.ProjectId
	testedZone = "asia-east1-b"

	log.Println("======== SetupSuite  ========")
}

func (suite *StorageManagerTestSuite) Test01_ListBuckets() {
	testedStorageManager.ListBuckets("livehouse-test")
}

func (suite *StorageManagerTestSuite) TearDownSuite() {
	log.Println("======== TearDown  ========")
}
