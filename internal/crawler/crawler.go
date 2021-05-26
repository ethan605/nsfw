package crawler

import (
	"fmt"
	"log"
	"time"

	"github.com/go-resty/resty/v2"
)

// Crawler represents a crawler instance
type Crawler interface {
	Start() error
	// Stop() error
	// Pause() error
}

// Profile provides information of a user
type Profile interface {
	fmt.Stringer
	AvatarURL() string
	DisplayName() string
	ID() string
	Username() string
}

// Writer provides interfaces to output profiles
type Writer interface {
	Write(profile Profile) error
}

// Config holds configurations for the crawler
type Config struct {
	// HTTP client, auto initialise with `resty.New()` if `nil`
	Client *resty.Client
	// The amount of time to wait between each request
	Defer time.Duration
	// Output pipeline
	Output Writer
	// The initial profile to start crawling with
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
