package crawler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

var (
	// SuggestedQueryHash is the query param to fetch suggested profiles
	SuggestedQueryHash = "d4d88dc1500312af6f937f7b804c68c3"

	// UserAgent header for crawler session
	UserAgent = "Mozilla/5.0 (X11; Linux x86_64; rv:88.0) Gecko/20100101 Firefox/88.0"
)

// NewInstagramCrawler initializes a crawler for instagram.com
func NewInstagramCrawler(config Config, scheduler Scheduler) (Crawler, error) {
	// TODO: spawn a headless browser, login & extract session ID from cookies
	const sessionID = "48056993126:Lyhz2Nv65JLp2E:26"

	httpClient := config.Client

	if httpClient == nil {
		httpClient = &http.Client{}
	}

	client := resty.NewWithClient(httpClient)

	if config.Seed == (Profile{}) {
		return nil, errors.New("missing required Seed config")
	}

	if config.Writer == nil {
		return nil, errors.New("missing required Writer config")
	}

	return &instagramSession{
		client:    client,
		config:    config,
		scheduler: scheduler,
		sessionID: sessionID,
	}, nil
}

// Run begins crawling data on instagram.com
func (s *instagramSession) Run() {
	go s.scheduler.Run(s.crawl, s.config.Seed)

	for profile := range s.scheduler.Results() {
		if err := s.config.Writer.Write(profile); err != nil {
			logrus.WithField("error", err).Error("writing failed")
		}
	}
}

/* Private stuffs */

var _ Crawler = (*instagramSession)(nil)

type instagramSession struct {
	client    *resty.Client
	config    Config
	scheduler Scheduler
	sessionID string
}

func (s *instagramSession) baseURL() string {
	return "https://www.instagram.com"
}

func (s *instagramSession) crawl(profile Profile) (Profile, []Profile, error) {
	profileDetail, err := s.fetchProfileDetail(profile)

	if err != nil {
		return Profile{}, nil, err
	}

	relatedProfiles, err := s.fetchRelatedProfiles(profile)

	if err != nil {
		return Profile{}, nil, err
	}

	return profileDetail, relatedProfiles, nil
}

func (s *instagramSession) fetchProfileDetail(profile Profile) (Profile, error) {
	type schema struct {
		Graphql struct {
			User instagramProfile
		}
	}

	resp, err := s.client.R().
		SetPathParam("username", profile.Username).
		SetQueryParam("__a", "1").
		SetHeader("User-Agent", UserAgent).
		SetCookie(&http.Cookie{
			Domain: ".instagram.com",
			Path:   "/",
			Name:   "sessionid",
			Value:  s.sessionID,
		}).
		SetResult(&schema{}).
		Get(s.baseURL() + "/{username}/")

	if err != nil {
		return Profile{}, err
	}

	if resp.StatusCode() != 200 {
		return Profile{}, errors.New("fetch profile error")
	}

	data, _ := resp.Result().(*schema)
	return data.Graphql.User.toProfile(), nil
}

func (s *instagramSession) fetchRelatedProfiles(fromProfile Profile) ([]Profile, error) {
	queryVariables := struct {
		UserID                 string `json:"user_id"`
		IncludeChaining        bool   `json:"include_chaining"`
		IncludeReel            bool   `json:"include_reel"`
		IncludeSuggestedUsers  bool   `json:"include_suggested_users"`
		IncludeLoggedOutExtras bool   `json:"include_logged_out_extras"`
		IncludeHighlightReels  bool   `json:"include_highlight_reels"`
		IncludeLiveStatus      bool   `json:"include_live_status"`
	}{
		UserID:                 fromProfile.ID,
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

	resp, err := s.client.R().
		SetQueryParams(map[string]string{
			"query_hash": SuggestedQueryHash,
			"variables":  string(variables),
		}).
		SetHeader("User-Agent", UserAgent).
		SetCookie(&http.Cookie{
			Name:  "sessionid",
			Value: s.sessionID,
		}).
		SetResult(&schema{}).
		Get(s.baseURL() + "/graphql/query")

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != 200 {
		return nil, errors.New("fetch related profiles error")
	}

	data, _ := resp.Result().(*schema)
	profiles := []Profile{}

	for _, edge := range data.Data.User.EdgeChaining.Edges {
		profiles = append(profiles, edge.Node.toProfile())
	}

	return profiles, nil
}

/* Private stuffs */

type instagramProfile struct {
	FullName      string `json:"full_name"`
	Username      string `json:"username"`
	ProfilePicURL string `json:"profile_pic_url_hd"`
	ID            string `json:"id"`
}

func (p instagramProfile) toProfile() Profile {
	return Profile{
		Source:      "Instagram",
		AvatarURL:   p.ProfilePicURL,
		DisplayName: p.FullName,
		ID:          p.ID,
		Username:    p.Username,
	}
}
