package crawler

import (
	"fmt"
	"log"
	"time"
)

// Crawler represents a crawler instance
type Crawler interface {
	Start()
	// Stop()
	// Pause()
}

// Profile provides information of a user
type Profile interface {
	fmt.Stringer
	AvatarURL() string
	DisplayName() string
	ID() string
	Username() string
}

// Config holds configurations for the crawler
type Config struct {
	// Defer is the amount of time to wait between each request
	Defer time.Duration
	// Seed contains information of the initial profile to start crawling with
	Seed Profile
}

/* Private stuffs */

type crawlSession interface {
	BaseURL() string
	FetchProfile() (Profile, error)
	FetchRelatedProfiles(fromProfile Profile) ([]Profile, error)
}

func panicOnError(err error) {
	if err != nil {
		log.Panicln(err)
	}
}
