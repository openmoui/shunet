//go:build windows
// +build windows

package utils

import (
	"os/exec"
	"strconv"
)

var logw = NewLogger()

func kill(pid int) {
	if pid <= 0 {
		logw.Fatalf("Cant kill shunet process with PID: %v", pid)
	}
	// Use taskkill command to terminate the process
	cmd := exec.Command("taskkill", "/F", "/PID", strconv.Itoa(pid))
	if err := cmd.Run(); err != nil {
		logw.Errorf("Error stopping shunet process PID:[%v], err : %+v", pid, err)
		return
	}
	logw.Infof("shunet Process with PID: %v has been stopped.", pid)
}
