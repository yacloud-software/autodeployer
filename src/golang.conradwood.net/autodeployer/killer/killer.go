package killer

import (
	"fmt"
	"golang.conradwood.net/go-easyops/linux"
	"syscall"
	"time"
)

func KillPID(pid int, brutal bool) {
	name := fmt.Sprintf("Pid #%d", pid)
	fmt.Printf("Killing process %s in status %s\n", name, pidStatus(pid))
	syscall.Kill(pid, syscall.SIGTERM)
	for i := 0; i < 10; i++ {
		fmt.Printf("Whilst killing process %s, status is still %s\n", name, pidStatus(pid))
		time.Sleep(time.Duration(1) * time.Second)
	}
	syscall.Kill(pid, syscall.SIGKILL)
}

func pidStatus(pid int) *linux.ProcessState {
	return linux.PidStatus(pid)

}
