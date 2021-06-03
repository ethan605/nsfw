package crawler

import (
	"fmt"

	"github.com/go-resty/resty/v2"
)

// Crawler represents a crawler instance
type Crawler interface {
	Run() error
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
	// Writer writing stream
	Writer Writer
	// The initial profile to start crawling with
	Seed Profile
}
