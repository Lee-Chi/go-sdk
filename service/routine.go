package service

import (
	"context"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"
)

var (
	pool map[string]*Routine
	mtx  sync.Mutex
)

func init() {
	pool = map[string]*Routine{}
}

func register(name string, routine *Routine) error {
	mtx.Lock()
	defer mtx.Unlock()

	if _, ok := pool[name]; ok {
		return fmt.Errorf("%s exist", name)
	}

	pool[name] = routine

	return nil
}

func unregister(name string) {
	mtx.Lock()
	defer mtx.Unlock()

	delete(pool, name)
}

const (
	Stop time.Duration = -1
)

func isActived(d time.Duration) bool {
	return d >= 0
}

type Func func() error

type DayAt struct {
	hour     int
	min      int
	location *time.Location
}

type Routine struct {
	name  string
	f     Func
	count int64

	err             error
	executeDuration time.Duration
	executeCount    int64

	cycleDuration time.Duration
	passDuration  time.Duration
	failDuration  time.Duration

	dayAts []DayAt
}

func NewRoutine(name string, f Func) *Routine {
	return &Routine{
		name:  name,
		f:     f,
		count: 0,

		err:             nil,
		executeDuration: 0,
		executeCount:    0,

		cycleDuration: Stop,
		passDuration:  Stop,
		failDuration:  Stop,

		dayAts: []DayAt{},
	}
}

// Count count is the number of times you want to be executed, 0 means unlimited
func (r *Routine) Count(count int64) *Routine {
	r.count = count
	return r
}

func (r *Routine) Repeat(duration time.Duration) *Routine {
	if !isActived(duration) {
		return r
	}

	r.cycleDuration = duration
	return r
}

func (r *Routine) RepeatByPassFail(passDuration time.Duration, failDurtaion time.Duration) *Routine {
	if !isActived(passDuration) && !isActived(failDurtaion) {
		return r
	}

	r.passDuration = passDuration
	r.failDuration = failDurtaion
	return r
}
func (r *Routine) EveryDayAt(hour int, min int, location *time.Location) *Routine {
	if hour > 24 {
		hour = 24
	} else if hour < 0 {
		hour = 0
	}

	if min > 59 {
		min = 59
	} else if min < 0 {
		min = 0
	}

	r.dayAts = append(r.dayAts, DayAt{
		hour:     hour,
		min:      min,
		location: location,
	})
	return r
}

func (r *Routine) execute() {
	if r.count > 0 {
		r.executeCount++
	}
	start := time.Now()
	r.err = r.f()
	end := time.Now()
	r.executeDuration = end.Sub(start)
}

func (r Routine) nextDuration() time.Duration {
	if r.count > 0 && r.count <= r.executeCount {
		return Stop
	}

	candicates := []time.Duration{}

	if isActived(r.cycleDuration) {
		cycleDuration := r.cycleDuration - r.executeDuration
		if cycleDuration < 0 {
			cycleDuration = 0
		}
		candicates = append(candicates, cycleDuration)
	}

	if isActived(r.passDuration) || isActived(r.failDuration) {
		duration := r.passDuration
		if r.err != nil {
			duration = r.failDuration
		}
		duration = duration - r.executeDuration
		if duration < 0 {
			duration = 0
		}
		candicates = append(candicates, duration)
	}

	if len(r.dayAts) > 0 {
		now := time.Now()
		minDuration := 24 * time.Hour
		for _, at := range r.dayAts {
			cur := now.In(at.location)
			dayAt := time.Date(cur.Year(), cur.Month(), cur.Day(), at.hour, at.min, 0, 0, at.location)
			duration := dayAt.Sub(now)
			if duration < 0 {
				duration += 24 * time.Hour
			}
			if duration < time.Second {
				continue
			}

			if minDuration > duration {
				minDuration = duration
			}
		}
		candicates = append(candicates, minDuration)
	}

	if len(candicates) == 0 {
		return Stop
	}

	sort.Slice(candicates, func(i, j int) bool {
		return candicates[i] < candicates[j]
	})

	return candicates[0]
}

// Go delay represents the time to delay the first execution. if the delay parameter is not given, it means execution according to routine schedule.
func (r *Routine) Go(delay ...time.Duration) error {
	if err := register(r.name, r); err != nil {
		return err
	}

	go func() {
		defer unregister(r.name)

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		nextDuation := r.nextDuration()
		if len(delay) > 0 {
			nextDuation = delay[0]
		}
		timer := time.NewTimer(nextDuation)

		done := false
		for {
			select {
			case <-timer.C:
				r.execute()

				nextDuration := r.nextDuration()

				if isActived(nextDuration) {
					timer.Reset(nextDuration)
				} else {
					done = true
				}
			case <-quit:
				done = true
			}

			if done {
				break
			}
		}
	}()

	return nil
}

func Wait(ctx context.Context) {
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
