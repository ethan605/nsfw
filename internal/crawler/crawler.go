package crawler

import (
	"encoding/json"
	"errors"
	"io"
	"log"
)

// Source contains information of a crawling source
type Source struct {
	Name    string
	RawData io.Reader
}

// Seed contains data for the first profile to start crawling
type Seed struct {
	Category string `json:"category"`
	Username string `json:"username"`
	UserID   string `json:"user_id"`
}

// Session provides methods to access a crawling source
type Session interface {
	BaseURL() string
	FetchProfile() string
	FetchRelatedProfiles() string
}

// Crawl checks and runs crawler against given source
func Crawl(source Source) error {
	switch source.Name {
	case "instagram":
		crawlInstagram(source.RawData)
		return nil
	default:
		return errors.New("Invalid source")
	}
}

/* Private stuffs */

func parseSeeds(rawData io.Reader) []Seed {
	byteValue, err := io.ReadAll(rawData)
	panicOnError(err)

	var seeds []Seed
	err = json.Unmarshal(byteValue, &seeds)
	panicOnError(err)

	return seeds
}

func panicOnError(err error) {
	if err != nil {
		log.Panicln(err)
	}
}
