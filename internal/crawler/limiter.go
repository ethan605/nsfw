package crawler

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

// Limiter manages rate limiting based on time, max takes threshold and max background workers
type Limiter interface {
	Done(int)
	Take() bool
	Wait()
}

// LimiterConfig contains configurations for a Limiter
type LimiterConfig struct {
	DeferTime  time.Duration
	MaxTakes   int
	MaxWorkers int
}

// NewLimiter creates a scheduler,
// with an amount of time to wait between each take and an upper limit of total takes
func NewLimiter(config LimiterConfig) Limiter {
	deferTime := config.DeferTime

	if deferTime == 0 {
		deferTime = time.Millisecond
	}

	maxWorkers := config.MaxWorkers

	if maxWorkers == 0 {
		maxWorkers = 1
	}

	l := &limiter{
		deferTime:  deferTime,
		maxTakes:   uint32(config.MaxTakes),
		maxWorkers: maxWorkers,

		throttle: make(chan struct{}),
		wg:       &sync.WaitGroup{},
	}

	l.newThrottle()

	return l
}

// Done increases the profiles counter by `delta`
func (l *limiter) Done(delta int) {
	atomic.AddUint32(&l.takesCounter, uint32(delta))
}

// Take blocks until the next allowing time,
// or return `false` if max profiles exceeded.
func (l *limiter) Take() bool {
	_, ok := <-l.throttle
	return ok
}

// Wait checks if the profiles counter didn't exceed `maxTakes`,
// then add up to gracefully exit the throttle goroutine
func (l *limiter) Wait() {
	atomicCounter := atomic.LoadUint32(&l.takesCounter)
	logrus.WithField("counter", atomicCounter).Debug("invoke")
	if atomicCounter < l.maxTakes {
		atomic.AddUint32(&l.takesCounter, l.maxTakes-atomicCounter)
	}
	l.wg.Wait()
}

/* Private stuffs */

type limiter struct {
	// Received configurations
	deferTime  time.Duration
	maxTakes   uint32
	maxWorkers int

	// takesCounter: atomic counter to keep track of crawled profiles
	// throttle: limit concurrent jobs by time and `maxTakes`
	// wg: wait for throttle goroutine to be done
	takesCounter uint32
	throttle     chan struct{}
	wg           *sync.WaitGroup
}

func (l *limiter) newThrottle() {
	l.throttle = make(chan struct{}, l.maxTakes)

	l.wg.Add(1)

	go func() {
		defer l.wg.Done()

		ticker := time.NewTicker(l.deferTime).C

		for {
			<-ticker
			atomicCounter := atomic.LoadUint32(&l.takesCounter)
			logrus.WithField("counter", atomicCounter).WithField("maxTakes", l.maxTakes).Debug("invoke throttle")

			if atomicCounter >= l.maxTakes {
				logrus.Debug("invoke inner")
				close(l.throttle)
				return
			}

			for worker := 0; worker < l.maxWorkers; worker++ {
				l.throttle <- struct{}{}
			}
		}
	}()
}
