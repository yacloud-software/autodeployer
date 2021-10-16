package main

import (
	l "golang.conradwood.net/go-easyops/linux"
	"golang.conradwood.net/go-easyops/utils"
)

func main() {
	TestExecuteContainer()
}

func TestExecuteContainer() {
	c := l.NewContainer()
	//	c.SetMemoryLimit(3000)
	c.SetTemplate("") // empty
	c.Execute([]string{"/bin/busybox", "sh"})
}

func copdir() {
	err := l.CopyDir("/tmp/x", "/tmp/y")
	utils.Bail("failed to copydir", err)
}
