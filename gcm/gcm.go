// Package gcm provides the basic APIs to communicate with Google cloud monitoring
package gcm

import (
	"fmt"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	monitor "google.golang.org/api/cloudmonitoring/v2beta2"
)

// BuildCloudMonitorService builds the singlton service for CloudMonitor
func BuildCloudMonitorService(serviceEmail string, key []byte) (*monitor.Service, error) {
	conf := &jwt.Config{
		Email:      serviceEmail,
		PrivateKey: key,
		Scopes: []string{
			monitor.MonitoringScope,
			monitor.CloudPlatformScope,
		},
		TokenURL: google.JWTTokenURL,
	}

	service, err := monitor.New(conf.Client(oauth2.NoContext))
	if err != nil {
		return nil, err
	}

	return service, nil
}

// GcmManager is for low level communication with Google CloudMonitor.
type GcmManager struct {
	*monitor.Service `inject:""`
}

// https://cloud.google.com/monitoring/v2beta2/timeseries/list
func (manager *GcmManager) GetAvgCpuUtilization(projectId, instanceName, interval string) (float64, error) {
	label := fmt.Sprintf("compute.googleapis.com/instance_name==%s", instanceName)
	metric := "compute.googleapis.com/instance/cpu/utilization"

	response, err := manager.Service.Timeseries.List(
		projectId, metric, time.Now().Format(time.RFC3339), nil).
		Labels(label).
		Timespan(interval).
		Window(interval).
		Aggregator("mean").
		Do()

	if err != nil {
		return 0, err
	}

	if len(response.Timeseries) == 0 {
		return 0, fmt.Errorf("No data")
	}

	utilization := *response.Timeseries[0].Points[0].DoubleValue
	return utilization, nil
}
