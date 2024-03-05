package service

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/Lee-Chi/go-sdk/logger"
)

var sig chan os.Signal

// Launch ready to start the service and awaiting the  signal to gracefully shut down
func Launch() {
	logger.Info("service is running...")

	sig = make(chan os.Signal, 1)

	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	<-sig
}
