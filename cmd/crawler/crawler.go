package main

import (
	"math/rand"
	"nsfw/internal/crawler"
	"runtime"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)

	crawlInstagram(true)
	expGoroutines()
}

/* Private stuffs */

type crawlerOutput struct{}

func (o *crawlerOutput) Write(profile crawler.Profile) error {
	logrus.Debug(" - Crawled profile:", profile)
	return nil
}

func crawlInstagram(dryRun bool) {
	if dryRun {
		return
	}

	// TODO: read from somewhere else
	seedProfile := "vox.ngoc.traan"

	config := crawler.Config{
		Defer:  2,
		Output: &crawlerOutput{},
		Seed: crawler.NewInstagramProfile(map[string]interface{}{
			"Username": seedProfile,
		}),
	}

	instagramCrawler, err := crawler.NewInstagramCrawler(config)
	panicOnError(err)

	err = instagramCrawler.Start()
	panicOnError(err)
}

func panicOnError(err error) {
	if err != nil {
		logrus.Panicln(err)
	}
}

func randomWait(min int) {
	rand.Seed(time.Now().UnixNano())
	time.Sleep((time.Duration)(rand.Intn(100)+min) * time.Millisecond)
}

func fetchProfile() int {
	randomWait(500)
	return 1
}

func fetchRelatedProfiles(fromProfile int) []int {
	randomWait(1000)
	profiles := []int{}

	for idx := 1; idx <= 3; idx++ {
		profiles = append(profiles, fromProfile*100+idx)
	}

	return profiles
}

type scheduler struct {
	wg       *sync.WaitGroup
	queue    chan int
	done     chan struct{}
	limitter <-chan time.Time
}

func (s *scheduler) writeProfile(profile int) {
	logrus.WithField("profile", profile).Debug("enqueue")
	s.queue <- profile
}

func (s *scheduler) run() {
	s.wg = &sync.WaitGroup{}

	seedProfile := fetchProfile()
	go s.writeProfile(seedProfile)

	s.wg.Add(1)
	go s.crawlProfiles(seedProfile, 0)

	s.wg.Wait()
	close(s.done)
}

func (s *scheduler) results() <-chan int {
	results := make(chan int)

	go func() {
		for {
			select {
			case profile := <-s.queue:
				results <- profile
			case <-s.done:
				close(results)
				return
			}
		}
	}()

	return results
}

func (s *scheduler) crawlProfiles(fromProfile int, level int) {
	if level >= 2 {
		s.wg.Done()
		return
	}

	<-s.limitter

	logrus.
		WithFields(logrus.Fields{"fromProfile": fromProfile, "time": time.Now().Format(time.RFC3339Nano)}).
		Debug("crawl")

	for _, profile := range fetchRelatedProfiles(fromProfile) {
		s.queue <- profile
		s.wg.Add(1)
		go s.crawlProfiles(profile, level+1)
	}

	s.wg.Done()
}

func expGoroutines() {
	queue := make(chan int)
	done := make(chan struct{})
	limitter := time.Tick(time.Second)

	s := scheduler{
		done:     done,
		limitter: limitter,
		queue:    queue,
	}

	go s.run()

	for profile := range s.results() {
		logrus.
			WithFields(logrus.Fields{"profile": profile}).
			Debug(" - write")
	}

	logrus.
		WithFields(logrus.Fields{"goroutines": runtime.NumGoroutine()}).
		Debug("Exitting")
}
