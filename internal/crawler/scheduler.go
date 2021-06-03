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

/* Private stuffs */

type crawlJob func(Profile) ([]Profile, error)

// NewScheduler creates a scheduler, with an amount of time to wait between each request
// and an upper limit of total profiles to be crawled
func NewScheduler(deferTime time.Duration, maxProfiles uint32) Scheduler {
	profilesQueue := make(chan Profile)
	wg := &sync.WaitGroup{}

	return &schedulerStruct{
		deferTime:     deferTime,
		maxProfiles:   maxProfiles,
		profilesQueue: profilesQueue,
		wg:            wg,
	}
}

type schedulerStruct struct {
	deferTime       time.Duration
	limiter         <-chan struct{}
	maxProfiles     uint32
	profilesCounter uint32
	profilesQueue   chan Profile
	wg              *sync.WaitGroup
}

func (s *schedulerStruct) Run(job crawlJob, profile Profile) {
	s.limiter = s.newLimiter()
	s.wg.Add(1)
	go s.runJob(job, profile)

	s.wg.Wait()
	close(s.profilesQueue)
}

func (s *schedulerStruct) Results() <-chan Profile {
	return s.profilesQueue
}

func (s *schedulerStruct) runJob(job crawlJob, profile Profile) {
	defer s.wg.Done()

	// This throttles jobs until `s.limiter` is closed
	_, ok := <-s.limiter

	if !ok {
		logrus.
			WithField("profile", profile).
			Info("Max profiles reached, stop crawling")
		return
	}

	profiles, err := job(profile)

	if err != nil {
		logrus.
			WithFields(logrus.Fields{
				"profile": profile,
				"error":   err,
			}).
			Error("Crawling profile error")
		return
	}

	numProfiles := len(profiles)
	atomic.AddUint32(&s.profilesCounter, (uint32)(numProfiles))

	s.wg.Add(numProfiles)
	for _, profile := range profiles {
		s.profilesQueue <- profile
		go s.runJob(job, profile)
	}
}

func (s *schedulerStruct) newLimiter() <-chan struct{} {
	limiter := make(chan struct{})

	go func() {
		ticker := time.NewTicker(s.deferTime).C

		for {
			<-ticker
			atomicCounter := atomic.LoadUint32(&s.profilesCounter)

			if atomicCounter >= s.maxProfiles {
				close(limiter)
				return
			}

			limiter <- struct{}{}
		}
	}()

	return limiter
}
