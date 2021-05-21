package crawler

import (
	"encoding/json"
	"io"
)

// Seed contains information of crawling sources
type Seed struct {
	Category string `json:"category"`
	Username string `json:"username"`
	UserID   string `json:"user_id"`
}

func parseSeeds(source io.Reader) []Seed {
	byteValue, err := io.ReadAll(source)
	panicOnError(err)

	var seeds []Seed
	err = json.Unmarshal(byteValue, &seeds)
	panicOnError(err)

	return seeds
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
