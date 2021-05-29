package crawler

import (
	"sync"
	"time"
)

// Scheduler runs crawling function and manages rate limiting, crawling depth, etc.
type Scheduler interface {
	Run(crawler crawlFunc, fromProfile Profile)
	Results() <-chan Profile
}

// NewScheduler creates a Scheduler-compatible instance
func NewScheduler(deferTime time.Duration, maxDepth int) Scheduler {
	queue := make(chan Profile)

	return &schedulerStruct{
		DeferTime: deferTime,
		MaxDepth:  maxDepth,
		queue:     queue,
	}
}

/* Private stuffs */

type crawlFunc func(Profile) []Profile

type schedulerStruct struct {
	DeferTime time.Duration
	MaxDepth  int
	limitter  <-chan time.Time
	queue     chan Profile
	wg        *sync.WaitGroup
}

func (s *schedulerStruct) Run(crawler crawlFunc, fromProfile Profile) {
	s.wg = &sync.WaitGroup{}

	s.wg.Add(1)
	go s.runCrawler(crawler, fromProfile, 0)

	s.wg.Wait()
	close(s.queue)
}

func (s *schedulerStruct) Results() <-chan Profile {
	return s.queue
}

func (s *schedulerStruct) runCrawler(crawler crawlFunc, fromProfile Profile, level int) {
	defer s.wg.Done()

	if level >= s.MaxDepth {
		return
	}

	<-s.limitter

	for _, profile := range crawler(fromProfile) {
		s.queue <- profile
		s.wg.Add(1)
		go s.runCrawler(crawler, profile, level+1)
	}
}
