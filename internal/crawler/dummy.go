package crawler

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

func NewDummyCrawler(config Config, limiterConfig LimiterConfig) (Crawler, error) {
	if config.Seed.Username == "" && config.Seed.ID == "" {
		return nil, errors.New("missing required Seed config")
	}

	if config.Writer == nil {
		return nil, errors.New("missing required Writer config")
	}

	return &dummySession{
		config:        config,
		limiterConfig: limiterConfig,
	}, nil
}

/* Private stuffs */

type dummySession struct {
	config        Config
	limiterConfig LimiterConfig
}

func (s *dummySession) Run() {
	limiter := NewLimiter(s.limiterConfig)
	profilesQueue := make(chan Profile)

	go func() {
		wg := &sync.WaitGroup{}

		wg.Add(1)
		go s.crawl(s.config.Seed, wg, limiter, profilesQueue)

		wg.Wait()
		limiter.Wait()

		close(profilesQueue)
	}()

	for profile := range profilesQueue {
		_ = s.config.Writer.Write(profile)
	}
}

func (s *dummySession) crawl(profile Profile, wg *sync.WaitGroup, limiter Limiter, profilesQueue chan<- Profile) {
	defer wg.Done()

	ok := limiter.Take()

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
	limiter.Done(1)

	relatedProfiles, err := s.fetchRelatedProfiles(profile)

	if err != nil {
		logrus.WithField("profile", profile).Error("fetchRelatedProfiles failed")
		return
	}

	wg.Add(len(relatedProfiles))

	for _, relatedProfile := range relatedProfiles {
		go s.crawl(relatedProfile, wg, limiter, profilesQueue)
	}
}

func (s *dummySession) fetchProfileDetail(profile Profile) (Profile, error) {
	if profile.ID == "-1/1" {
		return Profile{}, errors.New("fake error")
	}

	return profile, nil
}

func (s *dummySession) fetchRelatedProfiles(fromProfile Profile) ([]Profile, error) {
	if strings.HasPrefix(fromProfile.ID, "-1/") {
		return nil, errors.New("fake error")
	}

	profiles := []Profile{}

	for idx := 1; idx <= 5; idx++ {
		relatedProfile := Profile{ID: fmt.Sprintf("%s/%d", fromProfile.ID, idx)}
		profiles = append(profiles, relatedProfile)
	}

	return profiles, nil
}
