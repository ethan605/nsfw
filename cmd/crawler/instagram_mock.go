package main

import (
	"errors"
	"fmt"
	"math/rand"
	"nsfw/internal/crawler"
	"time"

	"github.com/sirupsen/logrus"
)

// MockInstagramCrawler initializes a mock crawler for instagram.com
func MockInstagramCrawler(config crawler.Config) (crawler.Crawler, error) {
	if config.Output == nil {
		return nil, errors.New("missing required Output config")
	}

	scheduler := crawler.NewScheduler(time.Second, 2)

	return &mockInstagramSession{
		Config:    config,
		scheduler: scheduler,
	}, nil
}

/* Private stuffs */

func randomWait(min int) {
	rand.Seed(time.Now().UnixNano())
	time.Sleep((time.Duration)(rand.Intn(100)+min) * time.Millisecond)
}

type mockInstagramSession struct {
	crawler.Config
	scheduler crawler.Scheduler
}

func (s *mockInstagramSession) Start() error {
	seedProfile := s.fetchProfile()
	go s.scheduler.Run(s.fetchRelatedProfiles, seedProfile)

	for profile := range s.scheduler.Results() {
		_ = s.Output.Write(profile)
	}

	return nil
}

func (s *mockInstagramSession) fetchProfile() crawler.Profile {
	randomWait(500)
	return crawler.NewInstagramSeed("1")
}

func (s *mockInstagramSession) fetchRelatedProfiles(fromProfile crawler.Profile) []crawler.Profile {
	logrus.
		WithFields(logrus.Fields{
			"profile": fromProfile.Username(),
			"time":    time.Now().Format("15:04:05.999"),
		}).
		Debug("crawling")

	profiles := []crawler.Profile{}

	for idx := 1; idx <= 3; idx++ {
		relatedProfile := crawler.NewInstagramSeed(fmt.Sprintf("%s/%d", fromProfile.Username(), idx))
		profiles = append(profiles, relatedProfile)
	}

	return profiles
}
