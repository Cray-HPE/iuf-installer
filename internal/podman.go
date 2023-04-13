package internal

//go:generate mockgen -destination=../mocks/internal/podman.go -package=mocks -source=podman.go
import (
	"fmt"

	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/k3d-io/k3d/v5/pkg/logger"
)

// ServiceName = "podman.socket"
const ServiceName = "podman.socket"

// PodmanService is the interface for the podman service
type PodmanService interface {
	PodmanInit() error
	isSocketServiceRunning() (bool, error)
	startSocketService() error
}

// podmanService is the implementation of the podman service
type podmanService struct {
	Dbus *dbus.Conn
}

// PodmanServiceInstance is the instance of the podman service
var PodmanServiceInstance PodmanService = newPodmanService()

func newPodmanService() PodmanService {
	conn, err := dbus.New()
	if err != nil {
		logger.Log().Fatalln(err)
	}
	return &podmanService{
		Dbus: conn,
	}
}

// PodmanInit checks if the podman socket service is running, if not, start it
func (p *podmanService) PodmanInit() error {
	defer p.Dbus.Close()

	isRunning, err := p.isSocketServiceRunning()
	if err != nil {
		logger.Log().Errorf("Error checking service status: %s\n", err)
		return err
	}
	if !isRunning {
		logger.Log().Infof("starting %s\n", ServiceName)
		err := p.startSocketService()
		if err != nil {
			logger.Log().Errorf("Error start service : %s, %s\n", ServiceName, err)
			return err
		}
	}
	logger.Log().Infof("%s is running\n", ServiceName)
	return nil
}

// isSocketServiceRunning checks if the podman socket service is running
func (p *podmanService) isSocketServiceRunning() (bool, error) {
	units, err := p.Dbus.ListUnits()
	if err != nil {
		return false, err
	}
	// Check if the service is running
	for _, unit := range units {
		if unit.Name == ServiceName {
			logger.Log().Debugf("unit: %s, state: %s, substate: %s\n", unit.Name, unit.ActiveState, unit.SubState)
			return unit.ActiveState == "active" && unit.SubState == "listening", nil
		}
	}

	return false, nil
}

// startSocketService starts the podman socket service
func (p *podmanService) startSocketService() error {
	ch := make(chan string)
	// Start the service
	_, err := p.Dbus.StartUnit(ServiceName, "replace", ch)
	if err != nil {
		return err
	}
	// Wait for the result
	result := <-ch
	if result != "done" {
		return fmt.Errorf("Failed to start service %s: %s", ServiceName, result)
	}

	return nil
}
