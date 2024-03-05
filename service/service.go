package service

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var sig chan os.Signal

func Wait(ctx context.Context) {
	sig = make(chan os.Signal, 1)

	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	<-sig

	timer := time.NewTimer(60 * time.Millisecond)
	for {
		select {
		case <-timer.C:
			if len(pool) == 0 {
				return
			}
			timer.Reset(time.Second)
		case <-ctx.Done():
			return
		}
	}
}

// Accept accepts the api with the given uuid
func Accept(uuid string) {
	register(uuid)
}

// Done stops the api with the given uuid
func Done(uuid string) time.Duration {
	return unregister(uuid)
}
