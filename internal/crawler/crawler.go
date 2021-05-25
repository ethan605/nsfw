package crawler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// HTTPClient abstracts a client to make HTTP requests
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Source contains information of a crawling source
type Source struct {
	Name    string
	RawData io.Reader
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

// Session provides methods to access a crawling source
type Session interface {
	BaseURL() string
	FetchProfile() (Profile, error)
	FetchRelatedProfiles(fromProfile Profile) ([]Profile, error)
}

/* Private stuffs */

type seedStruct struct {
	Category string `json:"category"`
	Username string `json:"username"`
	UserID   string `json:"user_id"`
}

func parseSeeds(rawData io.Reader) []seedStruct {
	byteValue, err := io.ReadAll(rawData)
	panicOnError(err)

	var seeds []seedStruct
	err = json.Unmarshal(byteValue, &seeds)
	panicOnError(err)

	return seeds
}

func panicOnError(err error) {
	if err != nil {
		log.Panicln(err)
	}
}
