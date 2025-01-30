package services

import (
	"errors"
	"log"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

func CreateService(serviceName string, driverPath string) (*mgr.Service, error) {

	SCM, err := mgr.Connect()

	if err != nil {
		return nil, err
	}

	service := CheckService(*SCM, serviceName)

	if service == nil {

		serviceConfig := mgr.Config{
			ServiceType:  windows.SERVICE_KERNEL_DRIVER,
			StartType:    windows.SERVICE_DEMAND_START,
			ErrorControl: windows.SERVICE_ERROR_IGNORE,
		}

		service, err := SCM.CreateService(serviceName, driverPath, serviceConfig)

		if err != nil {
			return nil, err
		}

		if !IsValidServiceConfig(service) {
			return nil, errors.New("invalid service config")
		}

		if err := service.Start(); err != nil {
			return nil, err
		}
	}
	return service, nil

}

func CheckRunning(service mgr.Service) {

	serviceState, err := service.Control(svc.Cmd(svc.Running))

	if err != nil {
		log.Panic(err)
	}

	if serviceState.State != svc.State(svc.Running) {
		service.Start()
	}
}

func CheckService(SCM mgr.Mgr, serviceName string) *mgr.Service {
	if service, err := SCM.OpenService(serviceName); err == nil {
		return service
	}
	return nil
}

func IsValidServiceConfig(service *mgr.Service) bool {
	serviceConfig, err := service.Config()

	if err != nil {
		return false
	}

	if serviceConfig.ServiceType != windows.SERVICE_KERNEL_DRIVER {
		return false
	}

	if serviceConfig.ErrorControl != windows.SERVICE_ERROR_IGNORE {
		return false
	}

	return true
}

func RemoveService(serviceName string) error {
	SCM, err := mgr.Connect()

	if err != nil {
		return err
	}

	service, err := SCM.OpenService(serviceName)

	if err != nil {
		return err
	}

	if _, err := service.Control(svc.Stop); err != nil {
		return err
	}

	if err := service.Delete(); err != nil {
		return err
	}

	return nil

}
