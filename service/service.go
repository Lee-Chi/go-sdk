package service

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	pool map[string]time.Time
	mtx  sync.Mutex
)

func init() {
	pool = map[string]time.Time{}
}

func register(name string) error {
	mtx.Lock()
	defer mtx.Unlock()

	if _, ok := pool[name]; ok {
		return fmt.Errorf("%s exist", name)
	}

	pool[name] = time.Now()

	return nil
}

func unregister(name string) time.Duration {
	mtx.Lock()
	defer mtx.Unlock()

	var duration time.Duration

	if found, ok := pool[name]; ok {
		duration = time.Now().Sub(found)
		delete(pool, name)
	}

	return duration
}

var sig chan os.Signal

func Wait(ctx context.Context) {
	sig = make(chan os.Signal, 1)

	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	fmt.Printf("[SERVICE] %s | service is running ...\n", time.Now().UTC().Format(time.DateTime))

	<-sig

	fmt.Printf("[SERVICE] %s | shut down signal received\n", time.Now().UTC().Format(time.DateTime))

	timer := time.NewTimer(60 * time.Millisecond)
	for {
		select {
		case <-timer.C:
			if len(pool) == 0 {
				return
			}

			fmt.Printf("[SERVICE] %s | waiting for %d active routines\n", time.Now().UTC().Format(time.DateTime), len(pool))

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
