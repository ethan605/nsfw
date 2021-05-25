package crawler

import (
	"encoding/json"
	"fmt"
	"io"
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
	Category() string
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

type crawlSeed struct {
	Category string `json:"category"`
	Username string `json:"username"`
	UserID   string `json:"user_id"`
}

func parseSeeds(rawData io.Reader) []crawlSeed {
	byteValue, err := io.ReadAll(rawData)
	panicOnError(err)

	var seeds []crawlSeed
	err = json.Unmarshal(byteValue, &seeds)
	panicOnError(err)

	return seeds
}

func panicOnError(err error) {
	if err != nil {
		log.Panicln(err)
	}
}
