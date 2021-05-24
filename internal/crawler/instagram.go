package crawler

import (
	"encoding/json"
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
		seedStruct: seeds[0],
		// sessionID:          "48056993126:ztGeXSRnmEZ6Z7:23",
		// suggestedQueryHash: "d4d88dc1500312af6f937f7b804",

		sessionID:          "227619971:4I93ZY9IQywppz:10",
		suggestedQueryHash: "d4d88dc1500312af6f937f7b804c68c3",
	}

	client := &http.Client{}
	session.FetchProfile(client)
	session.FetchRelatedProfiles(client, "201057170")
}

/* Private stuffs */

var _ Session = (*instagramSession)(nil)

type instagramSession struct {
	seedStruct
	sessionID string
	// The query_hash to query suggested users
	suggestedQueryHash string
}

func (s *instagramSession) BaseURL() string {
	return "https://instagram.com"
}

func (s *instagramSession) FetchProfile(client HttpClient) string {
	profileURL := fmt.Sprintf("%s/%s/?__a=1", s.BaseURL(), s.Username)
	resp := s.makeRequest(client, profileURL, nil)

	var data struct {
		Graphql struct {
			User instagramProfile
		}
	}

	err := json.Unmarshal(resp, &data)
	panicOnError(err)
	log.Println("Resp", data.Graphql.User)

	return string(resp)
}

func (s *instagramSession) FetchRelatedProfiles(client HttpClient, fromProfile string) string {
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
		UserID:                 fromProfile,
		IncludeChaining:        true,
		IncludeReel:            false,
		IncludeSuggestedUsers:  true,
		IncludeLoggedOutExtras: false,
		IncludeHighlightReels:  false,
		IncludeLiveStatus:      false,
	}

	variables, _ := json.Marshal(queryVariables)

	params := map[string]string{
		"query_hash": s.suggestedQueryHash,
		"variables":  string(variables),
	}
	resp := s.makeRequest(client, graphqlURL, params)

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
	panicOnError(err)
	log.Println("Resp:", data.Data.User.EdgeChaining.Edges)

	return string(resp)
}

/* Private stuffs */

func (s *instagramSession) makeRequest(client HttpClient, url string, params map[string]string) []byte {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	panicOnError(err)

	req.Header.Add("Cookie", fmt.Sprintf("sessionid=%s", s.sessionID))

	if params != nil {
		q := req.URL.Query()

		for key, value := range params {
			q.Add(key, value)
		}

		req.URL.RawQuery = q.Encode()
	}

	resp, err := client.Do(req)
	panicOnError(err)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	panicOnError(err)
	// log.Printf("body: %+v\n", string(body))

	return body
}

type instagramProfile struct {
	FullName      string `json:"full_name"`
	IgName        string `json:"username"`
	IsVerified    bool   `json:"is_verified"`
	ProfilePicURL string `json:"profile_pic_url_hd"`
	UserID        string `json:"id"`
}

var _ Profile = (*instagramProfile)(nil)

func (p *instagramProfile) AvatarURL() string   { return p.ProfilePicURL }
func (p *instagramProfile) DisplayName() string { return p.FullName }
func (p *instagramProfile) Username() string    { return p.IgName }
func (p *instagramProfile) ID() string          { return p.UserID }
func (p instagramProfile) String() string {
	return fmt.Sprintf("<Instagram %s %s (%s)>", p.UserID, p.IgName, p.FullName)
}
