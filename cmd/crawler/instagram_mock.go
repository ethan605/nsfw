package main

import (
	"errors"
	"fmt"
	"math/rand"
	"nsfw/internal/crawler"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

func mockInstagramCrawler(config crawler.Config, limiter crawler.Limiter) (crawler.Crawler, error) {
	if config.Writer == nil {
		return nil, errors.New("missing required Writer config")
	}

	return &mockInstagramSession{
		config:  config,
		limiter: limiter,
	}, nil
}

/* Private stuffs */

func randomWait(min int) {
	rand.Seed(time.Now().UnixNano())
	time.Sleep((time.Duration)(rand.Intn(100)+min) * time.Millisecond)
}

type mockInstagramSession struct {
	config  crawler.Config
	limiter crawler.Limiter
}

func (s *mockInstagramSession) Run() {
	seedProfile := crawler.Profile{ID: "1"}

	profilesQueue := make(chan crawler.Profile)

	go func() {
		wg := &sync.WaitGroup{}

		wg.Add(1)
		go s.crawl(seedProfile, wg, profilesQueue)

		wg.Wait()
		s.limiter.Wait()

		close(profilesQueue)
	}()

	for profile := range profilesQueue {
		_ = s.config.Writer.Write(profile)
	}
}

func (s *mockInstagramSession) crawl(profile crawler.Profile, wg *sync.WaitGroup, profilesQueue chan<- crawler.Profile) {
	defer wg.Done()

	ok := s.limiter.Take()

	if !ok {
		logrus.WithField("profile", profile).Info("max takes reached")
		return
	}

	logrus.WithFields(logrus.Fields{
		"profile": profile,
		"time":    time.Now().Format("15:04:05.000"),
	}).Info("crawling")

	profileDetail, err := s.fetchProfileDetail(profile)

	if err != nil {
		logrus.WithField("profile", profile).Error("fetchProfileDetail failed")
		return
	}

	profilesQueue <- profileDetail
	s.limiter.Done(1)

	relatedProfiles, err := s.fetchRelatedProfiles(profile)

	if err != nil {
		logrus.WithField("profile", profile).Error("fetchRelatedProfiles failed")
		return
	}

	wg.Add(len(relatedProfiles))

	for _, relatedProfile := range relatedProfiles {
		go s.crawl(relatedProfile, wg, profilesQueue)
	}
}

func (s *mockInstagramSession) fetchProfileDetail(profile crawler.Profile) (crawler.Profile, error) {
	randomWait(500)
	return profile, nil
}

func (s *mockInstagramSession) fetchRelatedProfiles(fromProfile crawler.Profile) ([]crawler.Profile, error) {
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
