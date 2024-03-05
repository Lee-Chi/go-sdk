package service

import (
	"fmt"
	"sync"
)

var sig chan os.Signal

// Launch ready to start the service and awaiting the  signal to gracefully shut down
func Launch() {
	fmt.Println("service is running...")
	
	sig = make(chan os.Signal, 1)

	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	
	<-sig
}