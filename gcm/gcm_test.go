package gcm_test

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"gogoo/config"
	"gogoo/gcm"

	"github.com/facebookgo/inject"
	"github.com/stretchr/testify/suite"
)

var testedCloudMonitorManager gcm.GcmManager
var testedProjectId string
var testedZone string
var vmName = flag.String("vm", "", "")

func TestCloudMonitorTestSuite(t *testing.T) {
	suite.Run(t, new(CloudMonitorTestSuite))
}

type CloudMonitorTestSuite struct {
	suite.Suite
}

func (suite *CloudMonitorTestSuite) SetupSuite() {
	gcloudConfig := config.LoadGcloudConfig(config.LoadAsset("/config/config.json"))
	key, _ := ioutil.ReadAll(config.LoadAsset("/config/key.pem"))

	// Construct dependency graph
	cloudmonitorService, _ := gcm.BuildCloudMonitorService(gcloudConfig.ServiceAccount, key)

	var g inject.Graph
	err := g.Provide(
		&inject.Object{Value: cloudmonitorService},
		&inject.Object{Value: &testedCloudMonitorManager},
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

// go test --env=Alpha --vm=zebra-rtc-alpha-as-496c0542-424f-44c6-8fda-98f3615ca320
func (suite *CloudMonitorTestSuite) Test01_GetAvgCpuUtilization() {
	if *vmName == "" {
		log.Printf("--vm flag not set")
		return
	}

	if utilization, err := testedCloudMonitorManager.GetAvgCpuUtilization(testedProjectId, *vmName, "10m"); err != nil {
		log.Printf("Error: %s", err)
	} else {
		log.Printf("utilization: %f", utilization)
	}
}

func (suite *CloudMonitorTestSuite) TearDownSuite() {
	log.Println("======== TearDown  ========")
}
