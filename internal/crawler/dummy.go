package crawler

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// NewDummyCrawler creates a new instance of DummyCrawler
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

var _ Crawler = (*dummySession)(nil)

type dummySession struct {
	// Received configurations
	config        Config
	limiterConfig LimiterConfig

	limiter       Limiter
	jobsWg        *sync.WaitGroup
	profilesQueue chan Profile
}

func (s *dummySession) Run() {
	s.limiter = NewLimiter(s.limiterConfig)
	s.profilesQueue = make(chan Profile)

	go func() {
		s.jobsWg = &sync.WaitGroup{}

		s.jobsWg.Add(1)
		go s.crawl(s.config.Seed)

		s.jobsWg.Wait()
		s.limiter.Wait()

		close(s.profilesQueue)
	}()

	for profile := range s.profilesQueue {
		_ = s.config.Writer.Write(profile)
	}
}

func (s *dummySession) crawl(profile Profile) {
	defer s.jobsWg.Done()

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

	s.profilesQueue <- profileDetail
	s.limiter.Done(1)

	relatedProfiles, err := s.fetchRelatedProfiles(profile)

	if err != nil {
		logrus.WithField("profile", profile).Error("fetchRelatedProfiles failed")
		return
	}

	s.jobsWg.Add(len(relatedProfiles))

	for _, relatedProfile := range relatedProfiles {
		go s.crawl(relatedProfile)
	}
}

func (s *dummySession) fetchProfileDetail(profile Profile) (Profile, error) {
	// time.Sleep(500)

	if profile.ID == "-1/1" {
		return Profile{}, errors.New("fake error")
	}

	return profile, nil
}

func (s *dummySession) fetchRelatedProfiles(fromProfile Profile) ([]Profile, error) {
	// time.Sleep(500)

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
