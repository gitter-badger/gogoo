// Package gogoo encapsulates google cloud api for more specific operation logic
package gogoo

import (
	"os"

	"github.com/browny/gogoo/cloudsql"
	"github.com/browny/gogoo/gce"
	"github.com/browny/gogoo/gcm"
	"github.com/browny/gogoo/gds"
	"github.com/browny/gogoo/replicapoolupdater"
	"github.com/facebookgo/inject"
)

var gogoo GoGoo
var gceManager gce.GceManager
var gdsManager gds.GdsManager
var gcmManager gcm.GcmManager
var cloudSqlManager cloudsql.CloudSqlManager
var rpuManager replicapoolupdater.RpuManager

// Input parameter object to initialize GoGoo
type AppContext struct {
	ServiceAccount      string
	KeyOfServiceAccount []byte
	ProjectId           string
}

// GoGoo acts as the handler to access different subpackages
type GoGoo struct {
	*gce.GceManager           `inject:""`
	*gds.GdsManager           `inject:""`
	*gcm.GcmManager           `inject:""`
	*cloudsql.CloudSqlManager `inject:""`
}

// Used to create a new GoGoo object.
func New(ctx AppContext) *GoGoo {
	buildDependencyGraph(ctx)

	return &gogoo
}

// Construct dependency graph
func buildDependencyGraph(ctx AppContext) {
	computeService, _ := gce.BuildGceService(ctx.ServiceAccount, ctx.KeyOfServiceAccount)
	_, client, _ := gds.BuildGdsContext(
		ctx.ServiceAccount,
		ctx.KeyOfServiceAccount,
		ctx.ProjectId)
	cloudmonitorService, _ := gcm.BuildCloudMonitorService(ctx.ServiceAccount, ctx.KeyOfServiceAccount)
	sqlService, _ := cloudsql.BuildCloudSqlService(ctx.ServiceAccount, ctx.KeyOfServiceAccount)
	rpuService, _ := replicapoolupdater.BuildRpuService(ctx.ServiceAccount, ctx.KeyOfServiceAccount)

	var g inject.Graph
	err := g.Provide(
		&inject.Object{Value: client},
		&inject.Object{Value: computeService},
		&inject.Object{Value: cloudmonitorService},
		&inject.Object{Value: sqlService},
		&inject.Object{Value: rpuService},
		&inject.Object{Value: &gdsManager},
		&inject.Object{Value: &gceManager},
		&inject.Object{Value: &gcmManager},
		&inject.Object{Value: &cloudSqlManager},
		&inject.Object{Value: &rpuManager},
		&inject.Object{Value: &gogoo},
	)
	if err != nil {
		os.Exit(1)
	}
	if err := g.Populate(); err != nil {
		os.Exit(1)
	}
}
