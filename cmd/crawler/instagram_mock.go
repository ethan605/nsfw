package main

import (
	"errors"
	"fmt"
	"math/rand"
	"nsfw/internal/crawler"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

func mockInstagramCrawler(config crawler.Config, scheduler crawler.Scheduler) (crawler.Crawler, error) {
	if config.Writer == nil {
		return nil, errors.New("missing required Writer config")
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

func (s *mockInstagramSession) Run() error {
	seedProfile := s.fetchProfile()
	_ = s.config.Writer.Write(seedProfile)

	go s.scheduler.Run(s.fetchRelatedProfiles, seedProfile)

	for profile := range s.scheduler.Results() {
		_ = s.config.Writer.Write(profile)
	}

	return nil
}

func (s *mockInstagramSession) fetchProfile() crawler.Profile {
	randomWait(500)
	return crawler.Profile{ID: "1"}
}

func (s *mockInstagramSession) fetchRelatedProfiles(fromProfile crawler.Profile) ([]crawler.Profile, error) {
	logrus.WithField("profile", fromProfile).Info("crawling")

	if strings.HasPrefix(fromProfile.ID, "-1/") {
		return nil, errors.New("fake error")
	}

	profiles := []crawler.Profile{}

	for idx := 1; idx <= 3; idx++ {
		relatedProfile := crawler.Profile{ID: fmt.Sprintf("%s/%d", fromProfile.ID, idx)}
		profiles = append(profiles, relatedProfile)
	}

	return profiles, nil
}
