package main

import (
	"fmt"
	"math/rand"
	"nsfw/internal/crawler"
	"runtime"
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
	seedInstagramUsername := "vox.ngoc.traan"

	config := crawler.Config{
		Output: &crawlerOutput{},
		Seed:   crawler.NewInstagramSeed(seedInstagramUsername),
	}

	instagramCrawler, err := crawler.NewInstagramCrawler(config)
	panicOnError(err)

	scheduler := crawler.NewScheduler(time.Second, 3)
	err = instagramCrawler.Start(scheduler)
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

type mockInstagramSession struct{}

func (s *mockInstagramSession) Start(scheduler crawler.Scheduler) {
	seedProfile := s.fetchProfile()
	go scheduler.Run(s.fetchRelatedProfiles, seedProfile)

	for profile := range scheduler.Results() {
		logrus.
			WithFields(logrus.Fields{"profile": profile.Username()}).
			Debug(" writing")
	}

	logrus.
		WithFields(logrus.Fields{"goroutines": runtime.NumGoroutine()}).
		Debug("Exitting")
}

func (m *mockInstagramSession) fetchProfile() crawler.Profile {
	randomWait(500)
	return crawler.NewInstagramSeed("1")
}

func (m *mockInstagramSession) fetchRelatedProfiles(fromProfile crawler.Profile) []crawler.Profile {
	logrus.
		WithFields(logrus.Fields{
			"profile": fromProfile.Username(),
			"time":    time.Now().Format("15:04:05.999"),
		}).
		Debug("crawling")

	// randomWait(200)
	profiles := []crawler.Profile{}

	for idx := 1; idx <= 3; idx++ {
		relatedProfile := crawler.NewInstagramSeed(fmt.Sprintf("%s/%d", fromProfile.Username(), idx))
		profiles = append(profiles, relatedProfile)
	}

	return profiles
}

func expGoroutines() {
	session := mockInstagramSession{}
	scheduler := crawler.NewScheduler(time.Second, 3)
	session.Start(scheduler)
}
