package utils

import (
	"syscall"

	"github.com/rabbitstack/fibratus/pkg/syscall/security"
)

func EnableSeDebugPrivilege() error {
	var token syscall.Token
	h, err := syscall.GetCurrentProcess()
	if err != nil {
		return err
	}
	err = syscall.OpenProcessToken(h, syscall.TOKEN_ADJUST_PRIVILEGES|syscall.TOKEN_QUERY, &token)
	if err != nil {
		return err
	}
	err = security.EnableTokenPrivileges(token, security.SeDebugPrivilege)
	if err != nil {
		return err
	}
	return nil
}
