package crawler

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"
)

// MockInstagramCrawler initializes a mock crawler for instagram.com
func MockInstagramCrawler(config Config) (Crawler, error) {
	if config.Output == nil {
		return nil, errors.New("missing required Output config")
	}

	scheduler := newScheduler(config.DeferTime, config.MaxProfiles)

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
	Config
	scheduler Scheduler
}

func (s *mockInstagramSession) Start() error {
	seedProfile := s.fetchProfile()
	go s.scheduler.Run(s.fetchRelatedProfiles, seedProfile)

	for profile := range s.scheduler.Results() {
		_ = s.Output.Write(profile)
	}

	return nil
}

func (s *mockInstagramSession) fetchProfile() Profile {
	randomWait(500)
	return instagramProfile{UserID: "1"}
}

func (s *mockInstagramSession) fetchRelatedProfiles(fromProfile Profile) []Profile {
	logrus.
		WithFields(logrus.Fields{
			"profile": fromProfile.ID(),
			"time":    time.Now().Format("15:04:05.999"),
		}).
		Debug("crawling")

	profiles := []Profile{}

	for idx := 1; idx <= 3; idx++ {
		relatedProfile := instagramProfile{UserID: fmt.Sprintf("%s/%d", fromProfile.ID(), idx)}
		profiles = append(profiles, relatedProfile)
	}

	return profiles
}
