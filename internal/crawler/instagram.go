package crawler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/go-resty/resty/v2"
)

// NewInstagramCrawler initializes a crawler for instagram.com
func NewInstagramCrawler(client *resty.Client, source io.Reader) Crawler {
	seeds := parseSeeds(source)

	return &instagramSession{
		RestyClient: resty.New(),
		Seed:        seeds[0],

		SessionID:          "48056993126:bT885uz5tm3eBr:22",
		SuggestedQueryHash: "d4d88dc1500312af6f937f7b804c68c3",
	}
}

// Crawl crawls data on instagram.com
func (s *instagramSession) Crawl() {
	seedProfile, err := s.FetchProfile()
	panicOnError(err)
	log.Println("Seed profile:", seedProfile)

	relatedProfiles, err := s.FetchRelatedProfiles(seedProfile)
	panicOnError(err)

	for _, profile := range relatedProfiles {
		log.Println("- Related profiles:", profile)
	}
}

/* Private stuffs */

var _ crawlSession = (*instagramSession)(nil)

type instagramSession struct {
	RestyClient *resty.Client
	Seed        crawlSeed

	// Cookie
	SessionID string
	// The query_hash to query suggested users
	SuggestedQueryHash string
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

	resp, err := s.RestyClient.R().
		SetPathParam("username", s.Seed.Username).
		SetQueryParam("__a", "1").
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

	profile := &data.Graphql.User
	profile.SeedCategory = s.Seed.Category

	return profile, nil
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

	resp, err := s.RestyClient.R().
		SetQueryParams(map[string]string{
			"query_hash": s.SuggestedQueryHash,
			"variables":  string(variables),
		}).
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
		return nil, errors.New("Fetch related profiles error")
	}

	data, _ := resp.Result().(*schema)
	profiles := []Profile{}

	for _, edge := range data.Data.User.EdgeChaining.Edges {
		edge.Node.SeedCategory = s.Seed.Category
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
	SeedCategory  string
}

var _ Profile = (*instagramProfile)(nil)

func (p instagramProfile) AvatarURL() string   { return p.ProfilePicURL }
func (p instagramProfile) Category() string    { return p.SeedCategory }
func (p instagramProfile) DisplayName() string { return p.FullName }
func (p instagramProfile) Username() string    { return p.IgName }
func (p instagramProfile) ID() string          { return p.UserID }
func (p instagramProfile) String() string {
	return fmt.Sprintf("<Instagram %s %s %s %s>", p.SeedCategory, p.UserID, p.IgName, p.FullName)
}
