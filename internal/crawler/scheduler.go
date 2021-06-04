package crawler

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

// Scheduler runs crawling function and manages rate limiting, crawling depth, etc.
type Scheduler interface {
	Run(job crawlJob, profile Profile)
	Results() <-chan Profile
}

type SchedulerConfig struct {
	DeferTime   time.Duration
	MaxProfiles int
	MaxWorkers  int
}

/* Private stuffs */

type crawlJob func(Profile) ([]Profile, error)

// NewScheduler creates a scheduler, with an amount of time to wait between each request
// and an upper limit of total profiles to be crawled
func NewScheduler(config SchedulerConfig) Scheduler {
	profilesQueue := make(chan Profile)
	jobsWg := &sync.WaitGroup{}
	limiterWg := &sync.WaitGroup{}

	deferTime := config.DeferTime

	if deferTime == 0 {
		deferTime = 10 * time.Millisecond
	}

	maxWorkers := config.MaxWorkers

	if maxWorkers == 0 {
		maxWorkers = 1
	}

	return &schedulerStruct{
		deferTime:   deferTime,
		maxProfiles: uint32(config.MaxProfiles),
		maxWorkers:  maxWorkers,

		jobsWg:        jobsWg,
		limiterWg:     limiterWg,
		profilesQueue: profilesQueue,
	}
}

type schedulerStruct struct {
	// Received configurations
	deferTime   time.Duration
	maxProfiles uint32
	maxWorkers  int

	// jobsWg: wait for all crawling jobs to be done
	// limiter: limit concurrent jobs by time and `maxProfiles`
	// limiterWg: wait for limiter goroutine to be done
	jobsWg          *sync.WaitGroup
	limiter         <-chan struct{}
	limiterWg       *sync.WaitGroup
	profilesCounter uint32
	profilesQueue   chan Profile
}

func (s *schedulerStruct) Run(job crawlJob, profile Profile) {
	s.limiter = s.newLimiter()

	s.jobsWg.Add(1)
	go s.runJob(job, profile)
	s.jobsWg.Wait()

	// Check if the profiles counter didn't exceed `maxProfiles`,
	// then add up to gracefully exit the limiter goroutine
	if atomicCounter := atomic.LoadUint32(&s.profilesCounter); atomicCounter < s.maxProfiles {
		atomic.AddUint32(&s.profilesCounter, s.maxProfiles-atomicCounter)
	}
	s.limiterWg.Wait()

	// Finally close profilesQueue to enable iterating `Results()` via `range`
	close(s.profilesQueue)
}

func (s *schedulerStruct) Results() <-chan Profile {
	return s.profilesQueue
}

func (s *schedulerStruct) runJob(job crawlJob, profile Profile) {
	defer s.jobsWg.Done()

	// This throttles jobs until `s.limiter` is closed
	_, ok := <-s.limiter

	if !ok {
		logrus.
			WithField("profile", profile).
			Info("max profiles reached, stop crawling")
		return
	}

	profiles, err := job(profile)

	if err != nil {
		logrus.
			WithFields(logrus.Fields{
				"profile": profile,
				"error":   err,
			}).
			Error("crawling profile error")
		return
	}

	numProfiles := len(profiles)
	s.jobsWg.Add(numProfiles)
	atomic.AddUint32(&s.profilesCounter, uint32(numProfiles))

	for _, profile := range profiles {
		s.profilesQueue <- profile
		go s.runJob(job, profile)
	}
}

func (s *schedulerStruct) newLimiter() <-chan struct{} {
	limiter := make(chan struct{})

	s.limiterWg.Add(1)
	go func() {
		defer s.limiterWg.Done()
		ticker := time.NewTicker(s.deferTime).C

		for {
			<-ticker
			atomicCounter := atomic.LoadUint32(&s.profilesCounter)

			if atomicCounter >= s.maxProfiles {
				close(limiter)
				return
			}

			for worker := 0; worker < s.maxWorkers; worker++ {
				limiter <- struct{}{}
			}
		}
	}()

	return limiter
}
