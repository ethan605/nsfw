package crawler

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

// CrawlInstagram crawls data from instagram.com
func crawlInstagram(source io.Reader) {
	seeds := parseSeeds(source)
	session := instagramSession{
		Seed:               seeds[0],
		sessionID:          "48056993126:ztGeXSRnmEZ6Z7:23",
		suggestedQueryHash: "d4d88dc1500312af6f937f7b804",
	}

	session.FetchProfile()
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
	profileURL := fmt.Sprintf("%s/%s/?__a=1", s.BaseURL(), s.Username)

	req, err := http.NewRequest(http.MethodGet, profileURL, nil)
	panicOnError(err)
	req.Header.Add("Cookie", fmt.Sprintf("sessionid=%s", s.sessionID))

	client := http.Client{}
	resp, err := client.Do(req)
	panicOnError(err)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	log.Printf("body: %+v\n", string(body))

	return profileURL
}

func (s *instagramSession) FetchRelatedProfiles() string {
	return fmt.Sprintf(
		"%s/?user_id=%s&session=%s&query_hash=%s",
		s.BaseURL(), s.UserID, s.sessionID, s.suggestedQueryHash,
	)
}
