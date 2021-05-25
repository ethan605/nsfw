package crawler

import (
	"fmt"
	"log"
)

// Crawler represents a crawler instance
type Crawler interface {
	Crawl()
}

// Profile provides information of a user
type Profile interface {
	fmt.Stringer
	AvatarURL() string
	DisplayName() string
	ID() string
	Username() string
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
