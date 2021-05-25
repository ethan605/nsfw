package crawler

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

// InstagramCrawler initializes a crawler for instagram.com
type InstagramCrawler struct {
	Client HTTPClient
	Source io.Reader
}

// Crawl crawls data on instagram.com
func (c *InstagramCrawler) Crawl() {
	seeds := parseSeeds(c.Source)
	session := instagramSession{
		Seed:   seeds[0],
		Client: c.Client,

		// SessionID:          "48056993126:ztGeXSRnmEZ6Z7:23",
		// SuggestedQueryHash: "d4d88dc1500312af6f937f7b804",

		SessionID:          "227619971:4I93ZY9IQywppz:10",
		SuggestedQueryHash: "d4d88dc1500312af6f937f7b804c68c3",
	}

	seedProfile, err := session.FetchProfile()
	panicOnError(err)
	log.Println("Seed profile:", seedProfile)

	relatedProfiles, err := session.FetchRelatedProfiles(seedProfile)
	panicOnError(err)

	for _, profile := range relatedProfiles {
		log.Println("- Related profiles:", profile)
	}
}

/* Private stuffs */

var _ Session = (*instagramSession)(nil)

type instagramSession struct {
	Seed   seedStruct
	Client HTTPClient

	// Cookie
	SessionID string
	// The query_hash to query suggested users
	SuggestedQueryHash string
}

func (s *instagramSession) BaseURL() string {
	return "https://instagram.com"
}

func (s *instagramSession) FetchProfile() (Profile, error) {
	profileURL := fmt.Sprintf("%s/%s/?__a=1", s.BaseURL(), s.Seed.Username)
	resp := s.makeRequest(profileURL, nil)

	var data struct {
		Graphql struct {
			User instagramProfile
		}
	}

	err := json.Unmarshal(resp, &data)

	if err != nil {
		return nil, err
	}

	profile := &data.Graphql.User
	profile.SeedCategory = s.Seed.Category

	return profile, nil
}

func (s *instagramSession) FetchRelatedProfiles(fromProfile Profile) ([]Profile, error) {
	graphqlURL := fmt.Sprintf("%s/graphql/query", s.BaseURL())

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

	params := map[string]string{
		"query_hash": s.SuggestedQueryHash,
		"variables":  string(variables),
	}
	resp := s.makeRequest(graphqlURL, params)

	var data struct {
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

	err := json.Unmarshal(resp, &data)

	if err != nil {
		return nil, err
	}

	profiles := []Profile{}

	for _, edge := range data.Data.User.EdgeChaining.Edges {
		edge.Node.SeedCategory = s.Seed.Category
		profiles = append(profiles, edge.Node)
	}

	return profiles, nil
}

/* Private stuffs */

func (s *instagramSession) makeRequest(url string, params map[string]string) []byte {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	panicOnError(err)

	req.Header.Add("Cookie", fmt.Sprintf("sessionid=%s", s.SessionID))

	if params != nil {
		q := req.URL.Query()

		for key, value := range params {
			q.Add(key, value)
		}

		req.URL.RawQuery = q.Encode()
	}

	resp, err := s.Client.Do(req)
	panicOnError(err)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	panicOnError(err)

	return body
}

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
