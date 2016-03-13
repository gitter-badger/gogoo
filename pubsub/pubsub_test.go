package pubsub_test

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"gogoo/config"
	"gogoo/pubsub"

	"github.com/facebookgo/inject"
	"github.com/stretchr/testify/suite"
)

var testedPbsbManager pubsub.PbsbManager
var testedProjectId string
var testedZone string

func TestPbsbManagerTestSuite(t *testing.T) {
	suite.Run(t, new(PbsbManagerTestSuite))
}

type PbsbManagerTestSuite struct {
	suite.Suite
}

func (suite *PbsbManagerTestSuite) SetupSuite() {
	gcloudConfig := config.LoadGcloudConfig(config.LoadAsset("/config/config.json"))
	key, _ := ioutil.ReadAll(config.LoadAsset("/config/key.pem"))

	// Construct dependency graph
	pbsbService, _ := pubsub.BuildPbsbService(gcloudConfig.ServiceAccount, key)

	var g inject.Graph
	err := g.Provide(
		&inject.Object{Value: pbsbService},
		&inject.Object{Value: &testedPbsbManager},
	)
	if err != nil {
		os.Exit(1)
	}
	if err := g.Populate(); err != nil {
		os.Exit(1)
	}
	// :~)

	testedPbsbManager.Setup()

	testedProjectId = gcloudConfig.ProjectId
	testedZone = "asia-east1-b"

	log.Println("======== SetupSuite  ========")
}

func (suite *PbsbManagerTestSuite) Test01_ListTopics() {
	testedPbsbManager.ListTopics("projects/livehouse-test")
}

func (suite *PbsbManagerTestSuite) TearDownSuite() {
	log.Println("======== TearDown  ========")
}
