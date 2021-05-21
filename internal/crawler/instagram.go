package crawler

import (
	"fmt"
	"io"
	"log"
)

// CrawlInstagram crawls data from instagram.com
func CrawlInstagram(source io.Reader) {
	seeds := parseSeeds(source)
	log.Printf("Seeds: %+v\n", seeds)
}

/* Private stuffs */

var _ Session = (*instagramSession)(nil)

type instagramSession struct {
	sessionID string
	// The query_hash to query suggested users
	suggestedQueryHash string
}

func (s *instagramSession) BaseURL() string {
	return "https://instagram.com"
}

func (s *instagramSession) ProfileURL(username string) string {
	return fmt.Sprintf("%s/%s/?__a=1", s.BaseURL(), username)
}
