package crawler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

var (
	// SeedProfile provides a default seed to start crawling with
	SeedProfile = instagramProfile{
		IgName: "vox.ngoc.traan",
		UserID: "48056993126",
	}

	// SuggestedQueryHash is the query param to fetch suggested profiles
	SuggestedQueryHash = "d4d88dc1500312af6f937f7b804c68c3"

	// UserAgent header for crawler session
	UserAgent = "Mozilla/5.0 (X11; Linux x86_64; rv:88.0) Gecko/20100101 Firefox/88.0"
)

// NewInstagramCrawler initializes a crawler for instagram.com
func NewInstagramCrawler(client *resty.Client, seed Profile) Crawler {
	// TODO: spawn a headless browser, login & extract session ID from cookies
	const sessionID = "48056993126:tcq6ZS8XmVd6uv:21"

	return &instagramSession{
		Client:    client,
		Seed:      seed,
		SessionID: sessionID,
	}
}

// Crawl crawls data on instagram.com
func (s *instagramSession) Crawl() {
	seedProfile, err := s.FetchProfile()
	panicOnError(err)
	log.Println("Seed profile:", seedProfile)

	relatedProfiles, err := s.FetchRelatedProfiles(seedProfile)
	panicOnError(err)

	resultsCount := 0
	uniqResults := map[string]Profile{}

	logProfile := func(profile Profile, level int) {
		switch level {
		case 1:
			log.Println(" - 1st level related profile:", profile)
		case 2:
			log.Println("    - 2nd level related profile:", profile)
		}

		resultsCount++
		uniqResults[profile.ID()] = profile
	}

	for idx, profile := range relatedProfiles {
		if idx >= 2 {
			break
		}

		logProfile(profile, 1)

		time.Sleep(1 * time.Second)

		pp, err := s.FetchRelatedProfiles(profile)
		panicOnError(err)

		for _, p := range pp {
			logProfile(p, 2)
		}
	}

	log.Println("resultsCount", resultsCount)
	log.Println("uniqResults count", len(uniqResults))
}

/* Private stuffs */

var _ crawlSession = (*instagramSession)(nil)

type instagramSession struct {
	Client *resty.Client
	Seed   Profile

	// Cookie
	SessionID string
}

func (s *instagramSession) BaseURL() string {
	return "https://www.instagram.com"
}

func (s *instagramSession) FetchProfile() (Profile, error) {
	type schema struct {
		Graphql struct {
			User instagramProfile
		}
	}

	resp, err := s.Client.R().
		SetPathParam("username", s.Seed.Username()).
		SetQueryParam("__a", "1").
		SetHeader("User-Agent", UserAgent).
		SetCookie(&http.Cookie{
			Name:  "sessionid",
			Value: s.SessionID,
		}).
		SetResult(&schema{}).
		Get(s.BaseURL() + "/{username}/")

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != 200 {
		return nil, errors.New("Fetch profile error")
	}

	data, _ := resp.Result().(*schema)
	return data.Graphql.User, nil
}

func (s *instagramSession) FetchRelatedProfiles(fromProfile Profile) ([]Profile, error) {
	queryVariables := struct {
		UserID                 string `json:"user_id"`
		IncludeChaining        bool   `json:"include_chaining"`
		IncludeReel            bool   `json:"include_reel"`
		IncludeSuggestedUsers  bool   `json:"include_suggested_users"`
		IncludeLoggedOutExtras bool   `json:"include_logged_out_extras"`
		IncludeHighlightReels  bool   `json:"include_highlight_reels"`
		IncludeLiveStatus      bool   `json:"include_live_status"`
	}{
		UserID:                 fromProfile.ID(),
		IncludeChaining:        true,
		IncludeReel:            false,
		IncludeSuggestedUsers:  true,
		IncludeLoggedOutExtras: false,
		IncludeHighlightReels:  false,
		IncludeLiveStatus:      false,
	}

	variables, _ := json.Marshal(queryVariables)

	type schema struct {
		Data struct {
			User struct {
				EdgeChaining struct {
					Edges []struct {
						Node instagramProfile
					}
				} `json:"edge_chaining"`
			}
		}
	}

	resp, err := s.Client.R().
		SetQueryParams(map[string]string{
			"query_hash": SuggestedQueryHash,
			"variables":  string(variables),
		}).
		SetHeader("User-Agent", UserAgent).
		SetCookie(&http.Cookie{
			Name:  "sessionid",
			Value: s.SessionID,
		}).
		SetResult(&schema{}).
		Get(s.BaseURL() + "/graphql/query")

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != 200 {
		log.Println("Body", string(resp.Body()))
		return nil, errors.New("Fetch related profiles error")
	}

	data, _ := resp.Result().(*schema)
	profiles := []Profile{}

	for _, edge := range data.Data.User.EdgeChaining.Edges {
		profiles = append(profiles, edge.Node)
	}

	return profiles, nil
}

/* Private stuffs */

type instagramProfile struct {
	FullName      string `json:"full_name"`
	IgName        string `json:"username"`
	ProfilePicURL string `json:"profile_pic_url_hd"`
	UserID        string `json:"id"`
}

var _ Profile = (*instagramProfile)(nil)

func (p instagramProfile) AvatarURL() string   { return p.ProfilePicURL }
func (p instagramProfile) DisplayName() string { return p.FullName }
func (p instagramProfile) Username() string    { return p.IgName }
func (p instagramProfile) ID() string          { return p.UserID }
func (p instagramProfile) String() string {
	return fmt.Sprintf("<Instagram %s %s %s>", p.UserID, p.IgName, p.FullName)
}
