package service

import (
	"fmt"
	"sync"
)

var sig chan os.Signal

func Start() {
	sig = make(chan os.Signal, 1)

	fmt.Println("service start")
}

func Wait() {
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	<-sig

	fmt.Println("service stop")
}
