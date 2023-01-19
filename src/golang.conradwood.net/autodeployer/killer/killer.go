package killer

import (
	"fmt"
	"golang.conradwood.net/go-easyops/linux"
	"syscall"
	"time"
)

func KillPID(ppid int, brutal bool) {
	var pids []int
	children, err := pidStatus(ppid).Children()
	if err != nil {
		fmt.Printf("failed to get children of pid %d: %s\n", ppid, err)
	} else {
		for _, c := range children {
			pids = append(pids, c.Pid())
		}
	}
	pids = append(pids, ppid)
	s := ""
	for _, pid := range pids {
		s = s + fmt.Sprintf("%d ", pid)
	}
	fmt.Printf("Killing %s pids\n", s)
	for _, pid := range pids {
		ps := pidStatus(pid)
		name := fmt.Sprintf("Pid #%d (%s)", pid, ps.Binary())
		if pid == 1 {
			panic("not killing init")
		}
		fmt.Printf("Killing process %s in status %s\n", name, ps.Status())
		syscall.Kill(pid, syscall.SIGTERM)
		done := false
		for i := 0; i < 10; i++ {
			if ps.Status() == linux.STATUS_STOPPED {
				done = true
				break
			}
			fmt.Printf("Whilst killing process %s, status is still %s\n", name, ps.Status())
			time.Sleep(time.Duration(1) * time.Second)
		}
		if !done {
			syscall.Kill(pid, syscall.SIGKILL)
		}
	}
}

func pidStatus(pid int) *linux.ProcessState {
	return linux.PidStatus(pid)

}
