package internal

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"strings"

	"github.com/k3d-io/k3d/v5/pkg/logger"
	"github.com/k3d-io/k3d/v5/pkg/types"
)

//go:generate mockgen -destination=../mocks/internal/image.go -package=mocks -source=image.go

// ImageService is the interface for the image service
type ImageService interface {
	LoadImages() error
	getImageNamesFromTarball() ([]string, error)
	checkIfImagesExist(images []string) (bool, error)
}

// imageService is the implementation of the image service
type imageService struct {
	podmanService PodmanService
}

func NewImageService(podmanService PodmanService) ImageService {
	return &imageService{
		podmanService: podmanService,
	}
}

func (i *imageService) LoadImages() error {
	logger.Log().Infof("Loading images")
	// get image names from tarball
	images, err := i.getImageNamesFromTarball()
	if err != nil {
		logger.Log().Errorf("Error getting image names from tarball: %s\n", err)
		return err
	}

	// check if images are already loaded (skip if force is set)
	imagesExist, err := i.checkIfImagesExist(images)
	if err != nil {
		logger.Log().Errorf("Error checking if images exist: %s\n", err)
		return err
	}
	if !imagesExist {
		// untar images
		// load images
	}
	return nil
}

// getImageNamesFromTarball gets the image names from the tarball
func (i *imageService) getImageNamesFromTarball() ([]string, error) {
	var res []string
	file, err := os.Open(AppConfig.Tarball)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// gzip reader
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		panic(err)
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)

	// iterate through the files in the archive
	for {
		header, err := tarReader.Next()

		if err != nil {
			if err == io.EOF {
				break // end of tar archive
			}
			panic(err)
		}
		if strings.Contains(header.Name, types.DefaultToolsImageRepo) && header.Typeflag == tar.TypeDir {
			imageNameWithTagAndSlash := strings.Split(header.Name, "/docker/")[1]
			imageNameWIthTag := strings.TrimRight(imageNameWithTagAndSlash, "")
			res = append(res, imageNameWIthTag)
		}
	}
	logger.Log().Infof("Images: %v\n", res)
	return res, nil
}

func (i *imageService) checkIfImagesExist(images []string) (bool, error) {
	if AppConfig.Force {
		return false, nil
	}

	// Establish a connection to the Podman service
	// ctx := context.Background()
	// conn, err := bindings.NewConnection(ctx, "unix:///run/podman/podman.sock")
	// if err != nil {
	// 	logger.Log().Errorf("Error connecting to Podman:", err)
	// 	return false, err
	// }
	// defer conn.Close()

	// for _, image := range images {
	// 	filter := fmt.Sprintf("reference=%s", image)
	// 	options := entities.ImageListOptions{
	// 		Filters: []string{filter},
	// 	}

	// 	imgList, err := images.List(ctx, conn, options)
	// 	if err != nil {
	// 		return false, err
	// 	}

	// 	return len(imgList) > 0, nil
	// }
	return false, nil
}
