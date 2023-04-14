package internal

import (
	"context"
	"os"

	"github.com/k3d-io/k3d/v5/cmd/cluster"
	k3cluster "github.com/k3d-io/k3d/v5/pkg/client"
	"github.com/k3d-io/k3d/v5/pkg/logger"
	"github.com/k3d-io/k3d/v5/pkg/runtimes"
	"github.com/k3d-io/k3d/v5/pkg/types"
)

//go:generate mockgen -destination=../mocks/internal/image.go -package=mocks -source=image.go

// K3dService is the interface for the image service
type K3dService interface {
	DeployK3d() error
}

// k3dService is the implementation of the image service
type k3dService struct {
	ctx context.Context
}

// NewK3dService creates a new image service
func NewK3dService(ctx context.Context) K3dService {
	return &k3dService{
		ctx: ctx,
	}
}

func (k *k3dService) DeployK3d() error {
	logger.Log().Infof("Deploying k3d")
	// set DOCKER_HOST=unix:///run/podman/podman.sock
	os.Setenv("DOCKER_HOST", "unix:///run/podman/podman.sock")

	if AppConfig.Force {
		k3cluster.ClusterDelete(
			k.ctx,
			runtimes.SelectedRuntime,
			&types.Cluster{Name: ClusterName},
			types.ClusterDeleteOpts{SkipRegistryCheck: true},
		)
	}

	clusters, err := k3cluster.ClusterList(k.ctx, runtimes.SelectedRuntime)
	if err != nil {
		logger.Log().Errorf("Error listing clusters: %s", err)
		return err
	}
	for _, cluster := range clusters {
		if cluster.Name == ClusterName {
			logger.Log().Infof("Cluster %s already exists", ClusterName)
			return nil
		}
	}

	k3dCmd := cluster.NewCmdClusterCreate()
	k3dCmd.SetArgs([]string{
		ClusterName,
		"--agents", "3",
		"--agents-memory", "4GB",
		"--servers", "1",
		"--servers-memory", "4GB",
		"--wait",
		"--no-lb",
		"--k3s-arg", "'--node-name=iuf-w001'@agent:0",
		"--k3s-arg", "'--node-name=iuf-w002'@agent:1",
		"--k3s-arg", "'--node-name=iuf-w003'@agent:2",
		"--k3s-arg", "'--node-name=iuf-m001'@server:0",
		"--k3s-arg", "'--snapshotter=native'@server:0",
		"--k3s-arg", "'--snapshotter=native'@agent:*",
		"--kubeconfig-update-default=false",
		"--api-port", "7443",
	})
	k3dCmd.Execute()

	return nil
}
