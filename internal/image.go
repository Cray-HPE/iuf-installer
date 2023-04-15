package internal

import (
	"os"
	"os/exec"
	"strings"

	"github.com/k3d-io/k3d/v5/pkg/logger"
)

//go:generate mockgen -destination=../mocks/internal/image.go -package=mocks -source=image.go

// ImageService is the interface for the image service
type ImageService interface {
	LoadImages() error
}

// imageService is the implementation of the image service
type imageService struct {
	podmanService PodmanService
}

// NewImageService creates a new image service
func NewImageService(podmanService PodmanService) ImageService {
	return &imageService{
		podmanService: podmanService,
	}
}

var k3dImages = []string{
	"artifactory.algol60.net/csm-docker/stable/ghcr.io/k3d-io/k3d-tools:5.4.9",
	"artifactory.algol60.net/csm-docker/stable/docker.io/rancher/k3s:v1.21.7-k3s1",
}

func (i *imageService) LoadImages() error {
	logger.Log().Infof("Loading images")
	// get untar targetDir
	targetDir := strings.Split(AppConfig.Tarball, "/")[len(strings.Split(AppConfig.Tarball, "/"))-1]
	targetDir = "/tmp/" + strings.Split(targetDir, ".tar")[0]
	logger.Log().Infof("Target dir: %s\n", targetDir)
	// check if targetDir exists or force is set
	if _, err := os.Stat(targetDir); os.IsNotExist(err) || AppConfig.Force {
		// remove targetDir
		err := os.RemoveAll(targetDir)
		if err != nil {
			logger.Log().Warnf("Error removing directory %s: %s\n", targetDir, err)
		}
		// untar
		cmd := exec.Command("tar", "xzf", AppConfig.Tarball, "-C", "/tmp")
		err = cmd.Run()
		if err != nil {
			logger.Log().Errorf("Error extracting %s: %s", AppConfig.Tarball, err)
			return err
		}
	}

	// load images
	for _, k3dImage := range k3dImages {
		logger.Log().Infof("Loading image %s from extracted file", k3dImage)
		cmd := exec.Command("tar", "-cf", targetDir+"/docker/"+k3dImage+".tar", targetDir+"/docker/"+k3dImage)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			logger.Log().Errorf("Error creating tarball for %s: %s", k3dImage, err)
			return err
		}

		logger.Log().Infof("Loading image %s into podman", k3dImage)
		cmd = exec.Command("podman", "load", "-i", targetDir+"/docker/"+k3dImage)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			logger.Log().Errorf("Error loading image %s: %s", k3dImage, err)
			return err
		}

		newTag := strings.Split(k3dImage, "stable/")[1]
		logger.Log().Infof("Re-tag image %s -> %s ", k3dImage, newTag)
		cmd = exec.Command("podman", "image", "tag", "localhost"+targetDir+"/docker/"+k3dImage, newTag)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			logger.Log().Errorf("Error re-tagging image %s: %s", k3dImage, err)
			return err
		}
		// make sure k3d is going to use our image
		os.Setenv("K3D_IMAGE_TOOLS", newTag)

		// load airgap images
		logger.Log().Infof("Loading airgap images")
		cmd = exec.Command("podman", "load", "-i", targetDir+"/docker/k3s-airgap-images-amd64.tar")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			logger.Log().Errorf("Error loading airgap images %s: %s", k3dImage, err)
			return err
		}

	}
	return nil
}
