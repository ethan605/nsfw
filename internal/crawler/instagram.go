package crawler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/go-resty/resty/v2"
)

var (
	// SuggestedQueryHash is the query param to fetch suggested profiles
	SuggestedQueryHash = "d4d88dc1500312af6f937f7b804c68c3"

	// UserAgent header for crawler session
	UserAgent = "Mozilla/5.0 (X11; Linux x86_64; rv:88.0) Gecko/20100101 Firefox/88.0"
)

// NewInstagramProfile creates a profile of an Instagram user
func NewInstagramProfile(data map[string]interface{}) Profile {
	avatarURL, _ := data["AvatarURL"].(string)
	displayName, _ := data["DisplayName"].(string)
	id, _ := data["ID"].(string)
	username, _ := data["Username"].(string)

	return instagramProfile{
		FullName:      displayName,
		IgName:        username,
		ProfilePicURL: avatarURL,
		UserID:        id,
	}
}

// NewInstagramCrawler initializes a crawler for instagram.com
func NewInstagramCrawler(config Config) (Crawler, error) {
	// TODO: spawn a headless browser, login & extract session ID from cookies
	const sessionID = "48056993126:Oy8vcmxDfwaQQ3:14"

	if config.Client == nil {
		config.Client = resty.New()
	}

	if config.Seed == nil {
		return nil, errors.New("missing required Seed config")
	}

	if config.Output == nil {
		return nil, errors.New("missing required Output config")
	}

	return &instagramSession{
		Config:    config,
		SessionID: sessionID,
	}, nil
}

// Start begins crawling data on instagram.com
func (s *instagramSession) Start() error {
	seedProfile, err := s.FetchProfile()

	if err != nil {
		return err
	}

	err = s.Config.Output.Write(seedProfile)

	if err != nil {
		return err
	}

	relatedProfiles, err := s.FetchRelatedProfiles(seedProfile)

	if err != nil {
		return err
	}

	for _, profile := range relatedProfiles {
		err = s.Config.Output.Write(profile)

		if err != nil {
			return err
		}
	}

	return nil
}

/* Private stuffs */

var _ Crawler = (*instagramSession)(nil)
var _ crawlSession = (*instagramSession)(nil)

type instagramSession struct {
	Config Config

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

	resp, err := s.Config.Client.R().
		SetPathParam("username", s.Config.Seed.Username()).
		SetQueryParam("__a", "1").
		SetHeader("User-Agent", UserAgent).
		SetCookie(&http.Cookie{
			Domain: ".instagram.com",
			Path:   "/",
			Name:   "sessionid",
			Value:  s.SessionID,
		}).
		SetResult(&schema{}).
		Get(s.BaseURL() + "/{username}/")

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != 200 {
		return nil, errors.New("fetch profile error")
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

	resp, err := s.Config.Client.R().
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
		return nil, errors.New("fetch related profiles error")
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
