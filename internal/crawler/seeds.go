package crawler

import (
	"io/fs"
	"log"
)

// Seed contains information of crawling sources
type Seed struct {
	Category string `json:"category"`
	Username string `json:"username"`
}

func handleErr(err error) {
	if err != nil {
		panic(err)
	}
}

func parseInstagramSeeds(fileSystem fs.FS) []Seed {
	byteValue, err := fs.ReadFile(fileSystem, "./configs/crawler/seeds/instagram.json")
	handleErr(err)

	log.Println("File content:", string(byteValue))

	var seeds []Seed
	// err = json.Unmarshal(byteValue, &seeds)
	// handleErr(err)

	return seeds
}
