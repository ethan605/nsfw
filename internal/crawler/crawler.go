package crawler

import (
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
)

// Crawler represents a crawler instance
type Crawler interface {
	Start() error
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
	Write(Profile) error
}

// Config holds configurations for the crawler
type Config struct {
	// HTTP client, auto initialise with `resty.New()` if `nil`
	Client *resty.Client
	// Amount of time to wait between each request
	DeferTime time.Duration
	// Upper limit of total profiles to be crawled
	MaxProfiles uint32
	// Output writing stream
	Output Writer
	// The initial profile to start crawling with
	Seed Profile
}
