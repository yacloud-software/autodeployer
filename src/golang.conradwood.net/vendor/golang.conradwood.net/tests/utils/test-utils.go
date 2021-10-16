package main

import (
	"fmt"
	"golang.conradwood.net/go-easyops/utils"
	"time"
)

func main() {
	print(14)
	print(65)
	print(125)
	print(120)
	print(0)
	secs := uint32(0)
	fmt.Printf("'not set' as Age: %s\n", utils.TimestampAgeString(secs))
}

func print(age int) {
	secs := uint32(time.Now().Unix()) - uint32(age)
	fmt.Printf("%d seconds as Age: %s\n", age, utils.TimestampAgeString(secs))
}
