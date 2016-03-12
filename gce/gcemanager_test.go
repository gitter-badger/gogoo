package gce_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"gogoo/config"
	"gogoo/gce"
	"gogoo/utility"

	"github.com/facebookgo/inject"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	compute "google.golang.org/api/compute/v1"
)

var testedGceManager gce.GceManager
var testedProjectId string
var testedZone string

func TestGceManagerTestSuite(t *testing.T) {
	suite.Run(t, new(GceManagerTestSuite))
}

type GceManagerTestSuite struct {
	suite.Suite
}

func (suite *GceManagerTestSuite) SetupSuite() {
	gcloudConfig := config.LoadGcloudConfig(config.LoadAsset("/config/config.json"))
	key, _ := ioutil.ReadAll(config.LoadAsset("/config/key.pem"))

	// Construct dependency graph
	computeService, _ := gce.BuildGceService(gcloudConfig.ServiceAccount, key)

	var g inject.Graph
	err := g.Provide(
		&inject.Object{Value: computeService},
		&inject.Object{Value: &testedGceManager},
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

func (suite *GceManagerTestSuite) Test01_NewVm() {
	template, _ := ioutil.ReadAll(config.LoadAsset("/config/instance_template.json"))
	vm, _ := testedGceManager.InitVmFromTemplate(template, "asia-east1-b")

	testedGceManager.NewVm(testedProjectId, testedZone, vm)
}

func (suite *GceManagerTestSuite) Test02_GetVm() {
	newVm, _ := testedGceManager.GetVm(testedProjectId, testedZone, "instance-test")

	result, _ := json.MarshalIndent(newVm, "", "  ")
	log.Println(string(result))
}

func (suite *GceManagerTestSuite) Test03_GetNatIP() {
	instance, _ := testedGceManager.GetVm(testedProjectId, testedZone, "instance-test")
	natIP := testedGceManager.GetNatIP(instance)

	log.Printf("NatIP: %s", natIP)
}

func (suite *GceManagerTestSuite) Test04_AttachTags() {
	testedGceManager.AttachTags(testedProjectId, testedZone, "instance-test", []string{"rtc-8000"})

	time.Sleep(3 * time.Second)
	vm, _ := testedGceManager.GetVm(testedProjectId, testedZone, "instance-test")
	assert.True(suite.T(), utility.InStringSlice(vm.Tags.Items, "rtc-8000"))
}

func (suite *GceManagerTestSuite) Test05_ListVms() {
	instanceList, _ := testedGceManager.ListVms(testedProjectId, testedZone)

	// List all instances
	for _, v := range instanceList.Items {
		result, _ := json.MarshalIndent(v, "", "  ")
		log.Println(string(result))
	}

	assert.NotNil(suite.T(), instanceList)
}

func (suite *GceManagerTestSuite) Test06_ListImages() {
	imageList, _ := testedGceManager.ListImages(testedProjectId)

	for _, v := range imageList.Items {
		result, _ := json.MarshalIndent(v, "", "  ")
		log.Println(string(result))
	}

	assert.NotNil(suite.T(), imageList)
}

func (suite *GceManagerTestSuite) Test07_NewDisk() {
	testedGceManager.NewDisk(testedProjectId, testedZone, "disk-test", "global/snapshots/snapshot-test", 20)
}

func (suite *GceManagerTestSuite) Test08_GetDisk() {
	disk, _ := testedGceManager.GetDisk(testedProjectId, testedZone, "disk-test")

	result, _ := json.MarshalIndent(disk, "", "  ")
	log.Println(string(result))

	assert.NotNil(suite.T(), disk)

	assert.Equal(suite.T(), "snapshot-test", testedGceManager.GetSnapshotOfDisk(disk))
}

func (suite *GceManagerTestSuite) Test09_GetSnapshots() {
	snapshots, _ := testedGceManager.GetSnapshots(testedProjectId)

	assert.Equal(suite.T(), 1, len(snapshots))
	assert.Equal(suite.T(), "snapshot-test", snapshots[0].Name)
}

func (suite *GceManagerTestSuite) Test10_GetLatestSnapshot() {
	testedSnapshots := []*compute.Snapshot{
		&compute.Snapshot{
			Name: "zebra-rtc-alpha-snapshot-201503032021"},
		&compute.Snapshot{
			Name: "zebra-rtc-alpha-snapshot-201502161516"},
		&compute.Snapshot{
			Name: "zebra-rtc-alpha-snapshot-201502131103"},
	}

	latestSnapshot, _ := testedGceManager.GetLatestSnapshot("alpha", testedSnapshots)

	assert.Equal(suite.T(), "zebra-rtc-alpha-snapshot-201503032021", latestSnapshot.Name)
}

func (suite *GceManagerTestSuite) Test11_StopVm() {
	var stoppedChecker = func(projectId, zone, instanceName string) (bool, error) {
		instanceStoppedObserver := make(chan bool)
		go testedGceManager.ProbeVmStopped(projectId, zone, instanceName, instanceStoppedObserver)

		done := <-instanceStoppedObserver
		if !done {
			return false, fmt.Errorf("VM not stopped: instance[%s]", instanceName)
		}
		return true, nil
	}

	testedGceManager.StopVm(testedProjectId, testedZone, "instance-test", stoppedChecker)

	// SetMachineType
	testedGceManager.SetMachineType(testedProjectId, testedZone, "instance-test", "f1-micro")
}

func (suite *GceManagerTestSuite) Test12_StartVm() {

	var preparedChecker = func(projectId, zone, instanceName string) (bool, error) {
		instanceRunningObserver := make(chan bool)
		go testedGceManager.ProbeVmRunning(projectId, zone, instanceName, instanceRunningObserver)

		if done := <-instanceRunningObserver; !done {
			return false, fmt.Errorf("VM not running")
		}
		return true, nil
	}

	testedGceManager.StartVm(testedProjectId, testedZone, "instance-test", preparedChecker)

	vm, _ := testedGceManager.GetVm(testedProjectId, testedZone, "instance-test")
	assert.True(suite.T(), strings.Contains(vm.MachineType, "f1-micro"))
}

func (suite *GceManagerTestSuite) Test13_GetSnapshot() {
	snapshot, _ := testedGceManager.GetSnapshot(testedProjectId, "snapshot-test")

	log.Printf("snapshot: %+v", snapshot)
}

func (suite *GceManagerTestSuite) Test14_ListDisks() {
	disks, _ := testedGceManager.ListDisks(testedProjectId, testedZone)

	log.Printf("disk: %+v", disks.Items[0])
}

func (suite *GceManagerTestSuite) Test15_DetachTags() {
	testedGceManager.DetachTags(testedProjectId, testedZone, "instance-test", []string{"rtc-8000"})

	time.Sleep(3 * time.Second)
	vm, _ := testedGceManager.GetVm(testedProjectId, testedZone, "instance-test")
	assert.False(suite.T(), utility.InStringSlice(vm.Tags.Items, "rtc-8000"))
}

func (suite *GceManagerTestSuite) TearDownSuite() {
	log.Println("======== TearDown  ========")

	testedGceManager.DeleteDisk(testedProjectId, testedZone, "disk-test")
}
