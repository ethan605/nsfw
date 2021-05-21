package crawler

import (
	"fmt"
	"io"
	"log"
)

// CrawlInstagram crawls data from instagram.com
func crawlInstagram(source io.Reader) {
	seeds := parseSeeds(source)
	log.Printf("Seeds: %+v\n", seeds)
}

/* Private stuffs */

var _ Session = (*instagramSession)(nil)

type instagramSession struct {
	Seed
	sessionID string
	// The query_hash to query suggested users
	suggestedQueryHash string
}

func (s *instagramSession) BaseURL() string {
	return "https://instagram.com"
}

func (s *instagramSession) FetchProfile() string {
	return fmt.Sprintf("%s/%s/?__a=1", s.BaseURL(), s.Username)
}

func (s *instagramSession) FetchRelatedProfiles() string {
	return fmt.Sprintf(
		"%s/?user_id=%s&session=%s&query_hash=%s",
		s.BaseURL(), s.UserID, s.sessionID, s.suggestedQueryHash,
	)
}
