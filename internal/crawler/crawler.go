package crawler

import (
	"fmt"
	"net/http"
)

// Crawler represents a crawler instance
type Crawler interface {
	Run()
}

// Profile provides information of a user
type Profile struct {
	fmt.Stringer
	Source      string
	AvatarURL   string
	DisplayName string
	Gallery     []string
	ID          string
	Username    string
}

func (p Profile) String() string {
	source := "Profile"

	if p.Source != "" {
		source = p.Source
	}

	return fmt.Sprintf("<%s %s %s %s>", source, p.ID, p.Username, p.DisplayName)
}

// Writer provides interfaces to output profiles
type Writer interface {
	Write(Profile) error
}

// Config holds configurations for the crawler
// @param Client: HTTP client, auto initialise with `resty.New()` if `nil`
// @param Seed: the initial profile to start crawling with
// @param SessionID: cookie session ID
// @param Writer: writing stream
type Config struct {
	Client    *http.Client
	Seed      Profile
	SessionID string
	Writer    Writer
}
