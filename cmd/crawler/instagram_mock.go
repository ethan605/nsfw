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

func (s *mockInstagramSession) Run() {
	seedProfile := crawler.Profile{ID: "1"}
	go s.scheduler.Run(s.crawl, seedProfile)

	for profile := range s.scheduler.Results() {
		_ = s.config.Writer.Write(profile)
	}
}

func (s *mockInstagramSession) crawl(profile crawler.Profile) (crawler.Profile, []crawler.Profile, error) {
	profileDetail, err := s.fetchProfileDetail(profile)

	if err != nil {
		return crawler.Profile{}, nil, err
	}

	relatedProfiles, err := s.fetchRelatedProfiles(profile)

	if err != nil {
		return crawler.Profile{}, nil, err
	}

	return profileDetail, relatedProfiles, nil
}

func (s *mockInstagramSession) fetchProfileDetail(profile crawler.Profile) (crawler.Profile, error) {
	randomWait(500)
	return profile, nil
}

func (s *mockInstagramSession) fetchRelatedProfiles(fromProfile crawler.Profile) ([]crawler.Profile, error) {
	logrus.WithFields(logrus.Fields{
		"profile": fromProfile,
		"time":    time.Now().Format("15:04:05.000"),
	}).Info("crawling")

	if strings.HasPrefix(fromProfile.ID, "-1/") {
		return nil, errors.New("fake error")
	}

	profiles := []crawler.Profile{}

	for idx := 1; idx <= 5; idx++ {
		relatedProfile := crawler.Profile{ID: fmt.Sprintf("%s/%d", fromProfile.ID, idx)}
		profiles = append(profiles, relatedProfile)
	}

	return profiles, nil
}
