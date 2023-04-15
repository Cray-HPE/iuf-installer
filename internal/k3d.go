package internal

import (
	"context"
	"os"
	"os/exec"

	"github.com/k3d-io/k3d/v5/cmd/cluster"
	k3cluster "github.com/k3d-io/k3d/v5/pkg/client"
	"github.com/k3d-io/k3d/v5/pkg/logger"
	"github.com/k3d-io/k3d/v5/pkg/runtimes"
	"github.com/k3d-io/k3d/v5/pkg/types"
	"github.com/sirupsen/logrus"
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
	os.Setenv("DOCKER_SOCK", "/run/podman/podman.sock")

	if AppConfig.Force {
		k3cluster.ClusterDelete(
			k.ctx,
			runtimes.SelectedRuntime,
			&types.Cluster{Name: ClusterName},
			types.ClusterDeleteOpts{SkipRegistryCheck: true},
		)
	}

	iufCluster, _ := k3cluster.ClusterGet(k.ctx, runtimes.Docker, &types.Cluster{Name: ClusterName})

	if iufCluster != nil {
		logger.Log().Infof("Cluster has been created already")
		return nil
	}
	logger.Logger.SetLevel(logrus.DebugLevel)
	// cp docker-init for podman
	// /usr/bin/docker-init
	cmd := exec.Command("mkdir", "-p", "/usr/libexec/podman")
	cmd.Run()
	cmd = exec.Command("cp", "/usr/bin/docker-init", "/usr/libexec/podman/catatonit")
	cmd.Run()

	k3dCmd := cluster.NewCmdClusterCreate()
	k3dCmd.SetArgs([]string{
		ClusterName,
		"-v", "/tmp/iuf-0.0.1-alpha.3/docker/k3s-airgap-images-amd64.tar:/var/lib/rancher/k3s/agent/images/k3s-airgap-images-amd64.tar",
		"--agents", "1",
		"--agents-memory", "4GB",
		"--servers", "1",
		"--servers-memory", "4GB",
		"--wait",
		"--no-lb",
		"--k3s-arg", "--node-name=iuf-w001@agent:0",
		// "--k3s-arg", "'--node-name=iuf-w002'@agent:1",
		// "--k3s-arg", "'--node-name=iuf-w003'@agent:2",
		"--k3s-arg", "--node-name=iuf-m001@server:0",
		"--k3s-arg", "--snapshotter=native@server:0",
		"--k3s-arg", "--snapshotter=native@agent:0",
		// "--k3s-arg", "--snapshotter=native@agent:1",
		// "--k3s-arg", "--snapshotter=native@agent:2",
		"--kubeconfig-update-default=false",
		"--network", "podman",
		"--api-port", "7443",
	})
	k3dCmd.Execute()

	return nil
}
