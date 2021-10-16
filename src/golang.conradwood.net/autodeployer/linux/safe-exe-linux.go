package linux

import (
	"flag"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"
)

var (
	cmdLock sync.Mutex
	curCmd  string
	LogExe  = flag.Bool("log_execute", false, "log execution of external commands")
)

// execute a command...
// print stdout/err (so it ends up in the logs)
// also we add a timeout - if program hangs we return an error
// rather than 'hanging' forever
// and we use a low-level lock to avoid calling binaries at the same time
func SafelyExecute(cmd []string) (string, error) {
	// avoid possible segfaults (afterall it's called 'safely...')
	if len(cmd) == 0 {
		return "", fmt.Errorf("no command specified for execute.")
	}
	if curCmd != "" {
		fmt.Printf("Waiting for %s to complete...\n", curCmd)
	}
	cmdLock.Lock()
	defer cmdLock.Unlock()
	curCmd = cmd[0]
	if curCmd == "sudo" {
		if len(curCmd) < 2 {
			return "", fmt.Errorf("sudo without parameters not allowed")
		}
		curCmd = cmd[1]
	}
	// execute
	if *LogExe {
		fmt.Printf("Executing %s\n", curCmd)
		fmt.Printf("Commandline: %s\n", strings.Join(cmd, " "))
	}
	c := exec.Command(cmd[0], cmd[1:]...)

	output, err := syncExecute(c, 5)
	if *LogExe {
		printOutput(curCmd, output)
	}
	curCmd = ""
	if err != nil {
		return "", err
	}
	return output, nil
}

// execute with timeout.
// sends SIGKILL to process on timeout and returns error
func syncExecute(c *exec.Cmd, timeout int) (string, error) {
	running := false
	killed := false
	timer1 := time.NewTimer(time.Second * time.Duration(timeout))
	go func() {
		<-timer1.C
		if running {
			c.Process.Kill()
			killed = true
		}
	}()
	// racecondition - timer might expire between
	// setting flag and starting process.
	// (if timer is really short)
	running = true
	b, err := c.CombinedOutput()
	running = false
	if killed {
		err = fmt.Errorf("Process killed after %d seconds\n", timeout)
	}
	return string(b), err
}

func printOutput(cmd string, output string) {
	fmt.Printf("====BEGIN OUTPUT OF %s====\n", cmd)
	fmt.Printf("%s\n", output)
	fmt.Printf("====END OUTPUT OF %s====\n", cmd)
}
