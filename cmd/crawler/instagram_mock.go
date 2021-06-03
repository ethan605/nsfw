package main

import (
	"errors"
	"fmt"
	"math/rand"
	"nsfw/internal/crawler"
	"time"

	"github.com/sirupsen/logrus"
)

func mockInstagramCrawler(config crawler.Config, scheduler crawler.Scheduler) (crawler.Crawler, error) {
	if config.Output == nil {
		return nil, errors.New("missing required Output config")
	}

	return &mockInstagramSession{
		config:    config,
		scheduler: scheduler,
	}, nil
}

/* Private stuffs */

func randomWait(min int) {
	rand.Seed(time.Now().UnixNano())
	time.Sleep((time.Duration)(rand.Intn(100)+min) * time.Millisecond)
}

type mockInstagramSession struct {
	config    crawler.Config
	scheduler crawler.Scheduler
}

func (s *mockInstagramSession) Start() error {
	seedProfile := s.fetchProfile()
	_ = s.config.Output.Write(seedProfile)

	go s.scheduler.Run(s.fetchRelatedProfiles, seedProfile)

	for profile := range s.scheduler.Results() {
		_ = s.config.Output.Write(profile)
	}

	return nil
}

func (s *mockInstagramSession) fetchProfile() crawler.Profile {
	randomWait(500)
	return crawler.NewInstagramSeed("1")
}

func (s *mockInstagramSession) fetchRelatedProfiles(fromProfile crawler.Profile) ([]crawler.Profile, error) {
	logrus.
		WithFields(logrus.Fields{
			"profile": fromProfile.Username(),
			"time":    time.Now().Format("15:04:05.999"),
		}).
		Debug("crawling")

	if fromProfile.Username() == "-1" {
		return nil, errors.New("fake error")
	}

	profiles := []crawler.Profile{}

	for idx := 1; idx <= 3; idx++ {
		relatedProfile := crawler.NewInstagramSeed(fmt.Sprintf("%s/%d", fromProfile.Username(), idx))
		profiles = append(profiles, relatedProfile)
	}

	return profiles, nil
}
