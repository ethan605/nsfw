package crawler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestNewInstagramCrawler(t *testing.T) {
	limiterConfig := LimiterConfig{}

	_, err := NewInstagramCrawler(Config{}, limiterConfig)
	assert.EqualError(t, err, "missing required Seed config")

	_, err = NewInstagramCrawler(Config{Seed: fakeProfile}, limiterConfig)
	assert.EqualError(t, err, "missing required Writer config")

	config := Config{
		Seed:   fakeProfile,
		Writer: &mockWriter{},
	}
	_, err = NewInstagramCrawler(config, limiterConfig)
	assert.Equal(t, nil, err)
}

func TestInstagramCrawlSuccess(t *testing.T) {
	// Setup HTTP requests mock
	client := &http.Client{}
	httpmock.ActivateNonDefault(client)
	defer httpmock.DeactivateAndReset()

	for _, id := range []string{"-1", "1234", "2345", "3456", "4567", "5678", "6789", "7890"} {
		profileFixture := generateProfileDetailFixture(id)
		profileResponder, _ := httpmock.NewJsonResponder(200, profileFixture)

		httpmock.RegisterResponder(
			"GET",
			fmt.Sprintf("/%s/?__a=1", "user_"+id),
			profileResponder,
		)
	}

	httpmock.RegisterResponder(
		"GET",
		"/graphql/query",
		func(req *http.Request) (*http.Response, error) {
			variables := object{}
			_ = json.Unmarshal([]byte(req.URL.Query().Get("variables")), &variables)
			userID, _ := variables["user_id"].(string)

			fixturesMap := object{
				"1234": generateRelatedProfilesFixture("2345", "3456", "4567", "5678", "-1"),
				"2345": generateRelatedProfilesFixture("3456", "4567", "5678", "6789"),
				"3456": generateRelatedProfilesFixture("4567", "5678", "6789", "7890"),
			}

			fixture := fixturesMap[userID]

			if fixture == nil {
				fixture = fixturesMap["1234"]
			}

			resp, _ := httpmock.NewJsonResponse(200, fixture)
			return resp, nil
		},
	)

	writer := &mockWriter{}
	config := Config{
		Client: client,
		Seed:   fakeProfile,
		Writer: writer,
	}
	limiterConfig := LimiterConfig{MaxTakes: 6}
	crawler, _ := NewInstagramCrawler(config, limiterConfig)

	crawler.Run()

	profileIDs := []string{}

	for _, profile := range writer.WrittenProfiles {
		profileIDs = append(profileIDs, profile.ID)
	}

	sort.Strings(profileIDs)
	assert.Equal(t, []string{"1234", "2345", "3456", "4567", "5678"}, profileIDs)
}

func TestInstagramCrawlFailure(t *testing.T) {
	client := &http.Client{}
	httpmock.ActivateNonDefault(client)
	defer httpmock.DeactivateAndReset()

	writer := &mockWriter{}
	config := Config{
		Client: client,
		Seed:   fakeProfile,
		Writer: writer,
	}
	limiterConfig := LimiterConfig{MaxTakes: 10}
	crawler, _ := NewInstagramCrawler(config, limiterConfig)

	// Fetch profile detail error
	assert.NotPanics(t, crawler.Run)

	profileResponder, _ := httpmock.NewJsonResponder(200, generateProfileDetailFixture(fakeID))
	httpmock.RegisterResponder(
		"GET",
		fmt.Sprintf("/%s/?__a=1", fakeProfile.Username),
		profileResponder,
	)

	// Fetch related profiles error
	assert.NotPanics(t, crawler.Run)
}

func TestFetchProfileDetail(t *testing.T) {
	client := &http.Client{}
	httpmock.ActivateNonDefault(client)
	defer httpmock.DeactivateAndReset()

	session := instagramSession{
		client: resty.NewWithClient(client),
	}

	// No profile detail responder error
	_, err := session.fetchProfileDetail(fakeProfile)
	assert.NotEqual(t, nil, err)

	httpmock.RegisterResponder(
		"GET",
		fmt.Sprintf("/%s/?__a=1", fakeProfile.Username),
		httpmock.NewStringResponder(500, "Invalid"),
	)

	_, err = session.fetchProfileDetail(fakeProfile)
	assert.EqualError(t, err, "fetch profile error")

	profileResponder, _ := httpmock.NewJsonResponder(200, generateProfileDetailFixture(fakeID))
	httpmock.RegisterResponder(
		"GET",
		fmt.Sprintf("/%s/?__a=1", fakeProfile.Username),
		profileResponder,
	)

	profileDetail, err := session.fetchProfileDetail(fakeProfile)
	assert.Equal(t, nil, err)
	assert.Equal(t, fakeID, profileDetail.ID)
	assert.Equal(t, "user_"+fakeID, profileDetail.Username)
	assert.Equal(t, "User "+fakeID, profileDetail.DisplayName)
}

func TestFetchRelatedProfiles(t *testing.T) {
	client := &http.Client{}
	httpmock.ActivateNonDefault(client)
	defer httpmock.DeactivateAndReset()

	session := instagramSession{
		client: resty.NewWithClient(client),
	}

	// No related profiles responder error
	_, err := session.fetchRelatedProfiles(fakeProfile)
	assert.NotEqual(t, nil, err)

	httpmock.RegisterResponder(
		"GET",
		"/graphql/query",
		httpmock.NewStringResponder(500, "Invalid"),
	)

	_, err = session.fetchRelatedProfiles(fakeProfile)
	assert.EqualError(t, err, "fetch related profiles error")

	relatedProfilesResponder, _ := httpmock.NewJsonResponder(200, generateRelatedProfilesFixture("2345", "3456", "4567", "5678"))
	httpmock.RegisterResponder(
		"GET",
		"/graphql/query",
		relatedProfilesResponder,
	)

	relatedProfiles, err := session.fetchRelatedProfiles(fakeProfile)
	assert.Equal(t, nil, err)
	assert.Equal(t, 4, len(relatedProfiles))
	assert.Equal(t, "2345", relatedProfiles[0].ID)
	assert.Equal(t, "3456", relatedProfiles[1].ID)
	assert.Equal(t, "4567", relatedProfiles[2].ID)
	assert.Equal(t, "5678", relatedProfiles[3].ID)
}

func TestInstagramSessions(t *testing.T) {
	session := instagramSession{}
	assert.Equal(t, "https://www.instagram.com", session.baseURL())
}

func TestInstagramProfile(t *testing.T) {
	profile := instagramProfile{}
	assert.Equal(t, []string{}, profile.toProfile().Gallery)

	fixture := `{
		"edge_owner_to_timeline_media": {
			"edges": [
				{"node": { "display_url": "fake-url-1" }},
				{"node": { "display_url": "fake-url-2" }},
				{"node": { "display_url": "fake-url-3" }}
			]
		}
	}`
	_ = json.Unmarshal([]byte(fixture), &profile)
	assert.Equal(t, []string{"fake-url-1", "fake-url-2", "fake-url-3"}, profile.toProfile().Gallery)
}

/* Private stuffs */

type object map[string]interface{}

var (
	fakeID      = "1234"
	fakeProfile = Profile{ID: fakeID, Username: "user_" + fakeID}
)

func generateProfileDetailFixture(id string) object {
	return object{
		"graphql": object{
			"user": generateProfileFixture(id),
		},
	}
}

func generateRelatedProfilesFixture(relatedIds ...string) object {
	edges := []object{}

	for _, id := range relatedIds {
		edges = append(edges, object{"node": generateProfileFixture(id)})
	}

	return object{
		"data": object{
			"user": object{
				"edge_chaining": object{
					"edges": edges,
				},
			},
		},
	}
}

func generateProfileFixture(id string) object {
	return object{
		"full_name":          "User " + id,
		"id":                 id,
		"profile_pic_url_hd": "https://profile-pic-url",
		"username":           "user_" + id,
	}
}
