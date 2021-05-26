package crawler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

type Object map[string]interface{}

var (
	fakeUsername   = "fake.user.name"
	profileFixture = Object{
		"graphql": Object{
			"user": Object{
				"full_name":          "Fake Name",
				"id":                 "1234",
				"profile_pic_url_hd": "https://profile-pic-url",
				"username":           fakeUsername,
			},
		},
	}
)

func TestNewInstagramProfile(t *testing.T) {
	profile := NewInstagramProfile(Object{})
	assert.Equal(t, "", profile.AvatarURL())
	assert.Equal(t, "", profile.DisplayName())
	assert.Equal(t, "", profile.ID())
	assert.Equal(t, "", profile.Username())

	profile = NewInstagramProfile(Object{
		"AvatarURL":   "https://fake-avatar-url",
		"DisplayName": "Fake Name",
		"ID":          "1234",
		"Username":    fakeUsername,
	})
	assert.Equal(t, "https://fake-avatar-url", profile.AvatarURL())
	assert.Equal(t, "Fake Name", profile.DisplayName())
	assert.Equal(t, "1234", profile.ID())
	assert.Equal(t, fakeUsername, profile.Username())
}

func TestNewInstagramCrawler(t *testing.T) {
	client := resty.New()
	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()

	seedProfile := instagramProfile{IgName: fakeUsername}
	crawler := NewInstagramCrawler(client, Config{Seed: seedProfile})

	profileResponder, _ := httpmock.NewJsonResponder(200, profileFixture)
	httpmock.RegisterResponder(
		"GET",
		fmt.Sprintf("/%s/?__a=1", seedProfile.Username()),
		profileResponder,
	)

	httpmock.RegisterResponder(
		"GET",
		"/graphql/query",
		func(req *http.Request) (*http.Response, error) {
			variables := Object{}
			_ = json.Unmarshal([]byte(req.URL.Query().Get("variables")), &variables)
			userID, _ := variables["user_id"].(string)

			fixturesMap := Object{
				"1234": generateRelatedProfilesFixture("2345", "3456", "4567", "5678"),
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

	assert.NotPanics(t, func() { crawler.Start() })
}

func TestInstagramSessions(t *testing.T) {
	session := instagramSession{}
	assert.Equal(t, "https://www.instagram.com", session.BaseURL())
}

func TestFetchProfileFails(t *testing.T) {
	client := resty.New()
	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()

	session := instagramSession{
		Client: client,
		Config: Config{
			Seed: instagramProfile{},
		},
	}

	_, err := session.FetchProfile()
	assert.NotEqual(t, nil, err)

	session = instagramSession{
		Client: client,
		Config: Config{
			Seed: instagramProfile{
				IgName: "invalid.user.name",
			},
		},
	}
	httpmock.RegisterResponder(
		"GET",
		"/invalid.user.name/?__a=1",
		httpmock.NewStringResponder(500, "Invalid"),
	)

	_, err = session.FetchProfile()
	assert.EqualError(t, err, "fetch profile error")
}

func TestFetchProfileSuccess(t *testing.T) {
	client := resty.New()
	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()

	session := instagramSession{
		Client: client,
		Config: Config{
			Seed: instagramProfile{
				IgName: fakeUsername,
			},
		},
	}

	fakeURL := fmt.Sprintf("/%s/?__a=1", fakeUsername)
	responder, _ := httpmock.NewJsonResponder(200, profileFixture)
	httpmock.RegisterResponder("GET", fakeURL, responder)

	profile, err := session.FetchProfile()
	assert.Equal(t, nil, err)
	assert.Equal(t, "https://profile-pic-url", profile.AvatarURL())
	assert.Equal(t, "1234", profile.ID())
	assert.Equal(t, "Fake Name", profile.DisplayName())
	assert.Equal(t, fakeUsername, profile.Username())
	assert.Equal(t, "<Instagram 1234 fake.user.name Fake Name>", profile.String())
}

func TestFetchRelatedProfilesFails(t *testing.T) {
	client := resty.New()
	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()

	session := instagramSession{
		Client: client,
	}

	_, err := session.FetchRelatedProfiles(instagramProfile{})
	assert.NotEqual(t, nil, err)

	session = instagramSession{
		Client: client,
		Config: Config{
			Seed: instagramProfile{
				IgName: "invalid.user.name",
			},
		},
	}
	httpmock.RegisterResponder(
		"GET",
		"/graphql/query",
		httpmock.NewStringResponder(500, "Invalid"),
	)

	_, err = session.FetchRelatedProfiles(instagramProfile{})
	assert.EqualError(t, err, "fetch related profiles error")
}

func TestFetchRelatedProfilesSuccess(t *testing.T) {
	client := resty.New()
	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()

	session := instagramSession{
		Client: client,
	}

	relatedProfilesFixture := generateRelatedProfilesFixture("2345", "3456", "4567", "5678")
	responder, _ := httpmock.NewJsonResponder(200, relatedProfilesFixture)
	httpmock.RegisterResponder("GET", "/graphql/query", responder)

	profiles, err := session.FetchRelatedProfiles(instagramProfile{})
	assert.Equal(t, nil, err)

	assert.Equal(t, 4, len(profiles))
	assert.Equal(t, "2345", profiles[0].ID())
	assert.Equal(t, "3456", profiles[1].ID())
	assert.Equal(t, "4567", profiles[2].ID())
	assert.Equal(t, "5678", profiles[3].ID())
}

/* Private stuffs */

func generateRelatedProfilesFixture(relatedIds ...string) Object {
	edges := []Object{}

	for _, id := range relatedIds {
		edges = append(edges, Object{"node": Object{"id": id}})
	}

	return Object{
		"data": Object{
			"user": Object{
				"edge_chaining": Object{
					"edges": edges,
				},
			},
		},
	}
}
