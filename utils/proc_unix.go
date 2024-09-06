//go:build !windows
// +build !windows

package utils

import (
	"log"
	"syscall"
)

func kill(pid int) {
	if pid <= 0 {
		log.Fatalf("Cant kill shunet process with PID: %v", pid)
	}
	// 发送终止信号
	if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
		Log.Errorf("Error stopping shunet process PID:[%v], err : %+v", pid, err)
		return
	}
	Log.Infof("shunet Process with PID: %v has been stopped.", pid)
}
