// Package gce communicates with compute engine
package gce

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/browny/gogoo/utility"

	log "github.com/cihub/seelog"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	compute "google.golang.org/api/compute/v1"
)

func gceError(errMessage string) error {
	return fmt.Errorf("GCE operation fails: %s", errMessage)
}

const (
	VmRunningTimeout    = 180 * time.Second
	VmStoppingTimeout   = 180 * time.Second
	DiskCreationTimeout = 180 * time.Second
)

type VmConditionChecker func(projectId, zone, instanceName string) (bool, error)

// BuildGceService builds the singlton service for GceManager
func BuildGceService(serviceEmail string, key []byte) (*compute.Service, error) {
	conf := &jwt.Config{
		Email:      serviceEmail,
		PrivateKey: key,
		Scopes: []string{
			compute.ComputeScope,
		},
		TokenURL: google.JWTTokenURL,
	}

	if service, err := compute.New(conf.Client(oauth2.NoContext)); err != nil {
		return nil, err
	} else {
		return service, nil
	}
}

// BySnapshotName is used to sort all gce snapshot by name
type BySnapshotName []*compute.Snapshot

func (a BySnapshotName) Len() int           { return len(a) }
func (a BySnapshotName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a BySnapshotName) Less(i, j int) bool { return a[i].Name < a[j].Name }

// GceManager is for low level communication with Google Compute Engine.
type GceManager struct {
	Service *compute.Service `inject:""`
}

// NewVm creates a new VM.
// This method block till the status of created VM is RUNNING
// or will be timeout if it takes over `VmRunningTimeout`.
// https://godoc.org/google.golang.org/api/compute/v1#InstancesService.Insert
func (manager *GceManager) NewVm(projectId, zone string, vm *compute.Instance) error {
	log.Tracef("New VM: project[%s], zone[%s]", projectId, zone)

	if _, err := manager.Service.Instances.Insert(projectId, zone, vm).Do(); err != nil {
		return gceError(err.Error())
	}

	// Pooling the status of the created vm
	vmRunningObserver := make(chan bool)
	go manager.ProbeVmRunning(projectId, zone, vm.Name, vmRunningObserver)

	done := <-vmRunningObserver
	if !done {
		return gceError(fmt.Sprintf("NewVM timeout: VM[]", vm.Name))
	}

	return nil
}

// GetVm gets a VM. If VM not existed, return nil.
// https://godoc.org/google.golang.org/api/compute/v1#InstancesService.Get
func (manager *GceManager) GetVm(projectId, zone, vmName string) (*compute.Instance, error) {
	log.Tracef("Get VM: project[%s], zone[%s], vmName[%s]", projectId, zone, vmName)

	if vm, err := manager.Service.Instances.Get(projectId, zone, vmName).Do(); err != nil {
		return nil, gceError(err.Error())
	} else {
		return vm, nil
	}
}

// DeleteVm deletes a VM.
// https://godoc.org/google.golang.org/api/compute/v1#InstancesService.Delete
func (manager *GceManager) DeleteVm(projectId, zone, vmName string) error {
	log.Tracef("Delete VM: project[%s], zone[%s], vmName[%s]", projectId, zone, vmName)

	if _, err := manager.Service.Instances.Delete(projectId, zone, vmName).Do(); err != nil {
		return gceError(err.Error())
	}

	return nil
}

// StopVm stops a VM.
// parameter `vcc` is the checker function to check if the VM is successfully stopped.
// https://godoc.org/google.golang.org/api/compute/v1#InstancesService.Stop
func (manager *GceManager) StopVm(projectId, zone, vmName string, vcc VmConditionChecker) (*compute.Operation, error) {
	log.Tracef("Stop instance: project[%s], zone[%s], vmName[%s]", projectId, zone, vmName)

	op, err := manager.Service.Instances.Stop(projectId, zone, vmName).Do()
	if err != nil {
		return nil, gceError(err.Error())
	}

	if pass, err := vcc(projectId, zone, vmName); pass {
		return op, nil
	} else {
		return nil, err
	}
}

// SetMachineType changes the machine type for a stopped instance to the machine type specified in the request.
// https://godoc.org/google.golang.org/api/compute/v1#InstancesService.SetMachineType
func (manager *GceManager) SetMachineType(projectId, zone, vmName, machineType string) (*compute.Operation, error) {
	log.Debugf("SetMachineType: project[%s], zone[%s], vmName[%s], type[%s]",
		projectId, zone, vmName, machineType)

	instanceService := compute.NewInstancesService(manager.Service)
	machineTypeUri := fmt.Sprintf("zones/%s/machineTypes/%s", zone, machineType)
	request := compute.InstancesSetMachineTypeRequest{MachineType: machineTypeUri}

	return instanceService.SetMachineType(projectId, zone, vmName, &request).Do()
}

// ResetInstance resets a instance.
// https://godoc.org/google.golang.org/api/compute/v1#InstancesService.Reset
func (manager *GceManager) ResetInstance(projectId, zone, vmName string) (*compute.Operation, error) {
	log.Debugf("Reset instance: project[%s], zone[%s], vmName[%s]", projectId, zone, vmName)

	return manager.Service.Instances.Reset(projectId, zone, vmName).Do()
}

// StartVm starts a VM.
// parameter `vcc` is the checker function to check if the VM is successfully started.
// https://godoc.org/google.golang.org/api/compute/v1#InstancesService.Start
func (manager *GceManager) StartVm(projectId, zone, vmName string, vcc VmConditionChecker) (*compute.Operation, error) {
	log.Tracef("Start instance: project[%s], zone[%s], vmName[%s]", projectId, zone, vmName)

	op, err := manager.Service.Instances.Start(projectId, zone, vmName).Do()
	if err != nil {
		return nil, gceError(err.Error())
	}

	if pass, err := vcc(projectId, zone, vmName); pass {
		return op, nil
	} else {
		return nil, err
	}
}

// ListVms lists all VMs.
// https://godoc.org/google.golang.org/api/compute/v1#InstancesService.List
func (manager *GceManager) ListVms(projectId, zone string) (*compute.InstanceList, error) {
	log.Tracef("List VMs: project[%s], zone[%s]", projectId, zone)

	res, err := manager.Service.Instances.List(projectId, zone).Do()
	if err != nil {
		return nil, gceError(err.Error())
	}

	return res, nil
}

// ListImages lists all images.
// https://godoc.org/google.golang.org/api/compute/v1#ImagesService.List
func (manager *GceManager) ListImages(projectId string) (*compute.ImageList, error) {
	log.Tracef("List images: project[%s]", projectId)

	res, err := manager.Service.Images.List(projectId).Do()
	if err != nil {
		return nil, gceError(err.Error())
	}

	return res, nil
}

// ListDisks lists all disks.
// https://godoc.org/google.golang.org/api/compute/v1#DisksService.List
func (manager *GceManager) ListDisks(projectId, zone string) (*compute.DiskList, error) {
	log.Tracef("List disks: project[%s], zone[%s]", projectId, zone)

	diskService := compute.NewDisksService(manager.Service)

	res, err := diskService.List(projectId, zone).Do()
	if err != nil {
		return nil, gceError(err.Error())
	}

	return res, nil
}

// NewDisk creates a new disk by specified snapshot.
// https://godoc.org/google.golang.org/api/compute/v1#DisksService.Insert
func (manager *GceManager) NewDisk(projectId, zone, name, sourceSnapshot string, sizeGb int64) error {
	log.Tracef("New disk: project[%s], zone[%s], name[%s], sourceSnapshot[%s]",
		projectId, zone, name, sourceSnapshot)

	diskService := compute.NewDisksService(manager.Service)

	disk := &compute.Disk{
		Name:           name,
		SizeGb:         sizeGb,
		SourceSnapshot: sourceSnapshot}

	if _, err := diskService.Insert(projectId, zone, disk).Do(); err != nil {
		return gceError(err.Error())
	}

	diskCreationObserver := make(chan bool)
	go manager.ProbeDiskCreation(projectId, zone, name, diskCreationObserver)

	done := <-diskCreationObserver
	if !done {
		return fmt.Errorf("NewDisk timeout: disk[%s]", name)
	}

	return nil
}

// GetDisk gets disk.
// https://godoc.org/google.golang.org/api/compute/v1#DisksService.Get
func (manager *GceManager) GetDisk(projectId, zone, diskName string) (*compute.Disk, error) {
	log.Tracef("Get disk: project[%s], zone[%s], diskName[%s]", projectId, zone, diskName)

	diskService := compute.NewDisksService(manager.Service)

	disk, err := diskService.Get(projectId, zone, diskName).Do()
	if err != nil {
		return nil, gceError(err.Error())
	}

	return disk, nil
}

// DeleteDisk deletes disk.
// https://godoc.org/google.golang.org/api/compute/v1#DisksService.Delete
func (manager *GceManager) DeleteDisk(projectId, zone, diskName string) error {
	log.Tracef("Delete disk: project[%s], zone[%s], diskName[%s]", projectId, zone, diskName)

	diskService := compute.NewDisksService(manager.Service)

	if _, err := diskService.Delete(projectId, zone, diskName).Do(); err != nil {
		return gceError(err.Error())
	}

	return nil
}

// GetSnapshots gets all snapshots of the project.
// https://godoc.org/google.golang.org/api/compute/v1#SnapshotsService.List
func (manager *GceManager) GetSnapshots(projectId string) ([]*compute.Snapshot, error) {
	log.Tracef("Get snapshots: project[%s]", projectId)

	snapshotService := compute.NewSnapshotsService(manager.Service)
	result, err := snapshotService.List(projectId).Do()
	if err != nil {
		return nil, gceError(err.Error())
	}

	snapshots := result.Items
	for _, snapshot := range snapshots {
		log.Tracef("snapshot: id[%d], name[%s]", snapshot.Id, snapshot.Name)
	}

	return snapshots, nil
}

// GetSnapshot gets the specific snapshot
// https://godoc.org/google.golang.org/api/compute/v1#SnapshotsService.Get
func (manager *GceManager) GetSnapshot(projectId, snapshot string) (*compute.Snapshot, error) {
	log.Tracef("Get snapshot: project[%s], snapshot[%s]", projectId, snapshot)

	snapshotService := compute.NewSnapshotsService(manager.Service)

	if result, err := snapshotService.Get(projectId, snapshot).Do(); err != nil {
		return nil, gceError(err.Error())
	} else {
		return result, nil
	}
}

// adjustTags adjusts tags of VM.
// https://godoc.org/google.golang.org/api/compute/v1#InstancesService.SetTags
func (manager *GceManager) adjustTags(
	projectId, zone, vmName string, tags []string, newTagsGenerator func([]string, []string) []string) (
	*compute.Operation, error) {

	vm, err := manager.GetVm(projectId, zone, vmName)
	if err != nil {
		return nil, err
	}

	vm.Tags.Items = newTagsGenerator(vm.Tags.Items, tags)

	op, err := manager.Service.Instances.SetTags(projectId, zone, vmName, vm.Tags).Do()
	if err != nil {
		return nil, gceError(err.Error())
	}

	return op, nil
}

// AttachTags attaches tags onto VM.
func (manager *GceManager) AttachTags(projectId, zone, vmName string, addedTags []string) (*compute.Operation, error) {
	log.Tracef("AttachTags: vm[%s], addedTags[%s]", vmName, addedTags)

	attacher := func(src, new []string) []string {
		for _, n := range new {
			src = append(src, n)
		}
		return src
	}

	return manager.adjustTags(projectId, zone, vmName, addedTags, attacher)
}

// DetachTags detaches tags from VM.
func (manager *GceManager) DetachTags(projectId, zone, vmName string, removedTages []string) (*compute.Operation, error) {
	log.Tracef("DetachTags: vm[%s], removedTages[%s]", vmName, removedTages)

	detacher := func(src, remove []string) []string {
		result := []string{}
		for _, s := range src {
			if utility.InStringSlice(remove, s) {
				continue
			}
			result = append(result, s)
		}
		return result
	}

	return manager.adjustTags(projectId, zone, vmName, removedTages, detacher)
}

// AddInstancesIntoInstanceGroup adds instances into some instance group
// https://godoc.org/google.golang.org/api/compute/v1#InstanceGroupsService.AddInstances
func (manager *GceManager) AddInstancesIntoInstanceGroup(
	projectId, zone, instanceGroupName string, instances []string) (
	*compute.Operation, error) {

	log.Tracef(
		"AddInstancesIntoInstanceGroup: project[%s], region[%s], instanceGroupName[%s], instances[%s]",
		projectId, zone, instanceGroupName, instances)

	instanceGroupService := compute.NewInstanceGroupsService(manager.Service)

	instanceReferences := []*compute.InstanceReference{}
	for _, instance := range instances {
		instanceRef := compute.InstanceReference{Instance: instance}
		instanceReferences = append(instanceReferences, &instanceRef)
	}
	request := compute.InstanceGroupsAddInstancesRequest{Instances: instanceReferences}

	op, err := instanceGroupService.AddInstances(projectId, zone, instanceGroupName, &request).Do()
	if err != nil {
		return nil, gceError(err.Error())
	}

	return op, nil
}

// GetLatestSnapshot gets latest snapshot with specified prefix in its name
func (manager *GceManager) GetLatestSnapshot(prefix string, snapshots []*compute.Snapshot) (*compute.Snapshot, error) {
	filteredSnapshots := []*compute.Snapshot{}
	for _, snapshot := range snapshots {
		if strings.Contains(snapshot.Name, prefix) {
			filteredSnapshots = append(filteredSnapshots, snapshot)
		}
	}

	sort.Sort(BySnapshotName(filteredSnapshots))
	if len(filteredSnapshots) < 1 {
		log.Warn("No snapshot found")
		return nil, fmt.Errorf("No snapshot found")
	}

	result := filteredSnapshots[len(filteredSnapshots)-1]
	log.Tracef("Latest snapshot found: name[%s]", result.Name)

	return result, nil
}

// InitVmFromTemplate builds the sample VM from template
func (manager *GceManager) InitVmFromTemplate(templateFile []byte, zone string) (*compute.Instance, error) {
	type TemplateParameter struct {
		Zone string
	}
	var tp = TemplateParameter{Zone: zone}

	tmpl, _ := template.New("test").Parse(string(templateFile[:]))
	var b bytes.Buffer
	tmpl.Execute(&b, tp)

	var vm compute.Instance
	err := json.Unmarshal(b.Bytes(), &vm)
	if err != nil {
		return nil, err
	}
	// :~)

	return &vm, nil
}

// ProbeVmRunning probes the VM status till its status is RUNNING or timeout
func (manager *GceManager) ProbeVmRunning(projectId, zone, vmName string, observer chan<- bool) {
	startTime := time.Now()

	for {
		if time.Now().Sub(startTime) > VmRunningTimeout {
			log.Warnf("VM creation Timeout: VM[%s]", vmName)
			observer <- false

			break
		}

		createdInstance, err := manager.GetVm(projectId, zone, vmName)

		if err != nil {
			log.Tracef("VM not yet Existed: VM[%s]", vmName)
			time.Sleep(10 * time.Second)

			continue
		}

		if createdInstance.Status != "RUNNING" {
			log.Tracef("VM not yet Running: VM[%s]", vmName)
			time.Sleep(10 * time.Second)

			continue
		}

		log.Infof("VM Running!: VM[%s]", vmName)
		observer <- true

		break
	}
}

// ProbeInstanceStopped probes the instance status till its status is Stopping or timeout
func (manager *GceManager) ProbeVmStopped(projectId, zone, vmName string, observer chan<- bool) {
	startTime := time.Now()

	for {
		if time.Now().Sub(startTime) > VmStoppingTimeout {
			log.Warnf("VM stop Timeout: VM[%s]", vmName)
			observer <- false

			break
		}

		createdInstance, err := manager.GetVm(projectId, zone, vmName)

		if err != nil {
			log.Tracef("VM not yet Existed: VM[%s]", vmName)
			time.Sleep(10 * time.Second)

			continue
		}

		if createdInstance.Status != "TERMINATED" {
			log.Tracef("VM not yet Stopped: VM[%s]", vmName)
			time.Sleep(10 * time.Second)

			continue
		}

		log.Infof("VM Stopped!: VM[%s]", vmName)
		observer <- true

		break
	}
}

// ProbeDiskCreation probes the disk status till its status is READY or timeout
func (manager *GceManager) ProbeDiskCreation(projectId, zone, diskName string, observer chan<- bool) {
	startTime := time.Now()

	for {
		if time.Now().Sub(startTime) > DiskCreationTimeout {
			log.Warnf("Disk creation Timeout: disk[%s]", diskName)
			observer <- false

			break
		}

		disk, err := manager.GetDisk(projectId, zone, diskName)
		if err != nil {
			log.Tracef("Disk not yet Created: name[%s]", diskName)
			time.Sleep(10 * time.Second)

			continue
		}
		if disk.Status != "READY" {
			log.Tracef("Disk not yet Ready: name[%s]", disk.Name)
			time.Sleep(5 * time.Second)

			continue
		}

		log.Infof("Disk Created!: name[%s]", disk.Name)
		observer <- true

		break
	}
}

// GetNatIP gets NAT IP address from VM
func (manager *GceManager) GetNatIP(vm *compute.Instance) string {
	if vm == nil {
		return "missing"
	}
	natIP := vm.NetworkInterfaces[0].AccessConfigs[0].NatIP

	log.Tracef("Got NatIP: VM[%s], ip[%s]", vm.Name, natIP)

	return natIP
}

// GetNetworkIP gets internal IP address from VM
func (manager *GceManager) GetNetworkIP(vm *compute.Instance) string {
	if vm == nil {
		return "missing"
	}
	networkIP := vm.NetworkInterfaces[0].NetworkIP

	log.Tracef("Got NetworkIP: VM[%s], ip[%s]", vm.Name, networkIP)

	return networkIP
}

// GetSnapshotOfDisk gets the snapshot name of the disk
func (manager *GceManager) GetSnapshotOfDisk(disk *compute.Disk) string {
	log.Tracef("Snapshot of the disk: snapshot[%s]", disk.SourceSnapshot)

	arr := strings.Split(disk.SourceSnapshot, "/")

	return arr[len(arr)-1]
}

func (manager *GceManager) PatchInstanceMachineType(machineType, targetType string) string {
	split := strings.Split(machineType, "/")
	split[len(split)-1] = targetType
	return strings.Join(split, "/")
}
