package crawler

import (
	"io"
	"log"
)

// Instagram crawls data from instagram.com
func Instagram(source io.Reader) {
	seeds := parseSeeds(source)
	log.Printf("Seeds: %+v\n", seeds)
}
