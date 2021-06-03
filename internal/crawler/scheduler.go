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
	queue := make(chan Profile)
	wg := &sync.WaitGroup{}
	limiter := time.NewTicker(deferTime).C

	return &schedulerStruct{
		limiter:     limiter,
		maxProfiles: maxProfiles,
		queue:       queue,
		wg:          wg,
	}
}

type schedulerStruct struct {
	limiter         <-chan time.Time
	maxProfiles     uint32
	profilesCounter uint32
	queue           chan Profile
	wg              *sync.WaitGroup
}

func (s *schedulerStruct) Run(job crawlJob, profile Profile) {
	s.wg.Add(1)
	go s.runJob(job, profile)

	s.wg.Wait()
	close(s.queue)
}

func (s *schedulerStruct) Results() <-chan Profile {
	return s.queue
}

func (s *schedulerStruct) runJob(job crawlJob, profile Profile) {
	defer s.wg.Done()

	<-s.limiter

	atomicCounter := atomic.LoadUint32(&s.profilesCounter)

	if atomicCounter >= s.maxProfiles {
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

	atomic.AddUint32(&s.profilesCounter, (uint32)(len(profiles)))

	for _, profile := range profiles {
		s.queue <- profile
		s.wg.Add(1)
		go s.runJob(job, profile)
	}
}
