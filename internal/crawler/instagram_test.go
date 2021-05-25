package crawler

import (
	"fmt"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

type Object map[string]interface{}

var (
	baseURL        = "https://www.instagram.com"
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
	relatedProfilesFixture = Object{
		"data": Object{
			"user": Object{
				"edge_chaining": Object{
					"edges": []Object{
						{"node": Object{"id": "1234"}},
						{"node": Object{"id": "2345"}},
						{"node": Object{"id": "3456"}},
						{"node": Object{"id": "4567"}},
					},
				},
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
		fmt.Sprintf("%s/%s/?__a=1", baseURL, seedProfile.Username()),
		profileResponder,
	)

	relatedProfilesResponder, _ := httpmock.NewJsonResponder(200, relatedProfilesFixture)
	httpmock.RegisterResponder(
		"GET",
		baseURL+"/graphql/query",
		relatedProfilesResponder,
	)

	assert.NotPanics(t, func() { crawler.Crawl() })
}

func TestInstagramSessions(t *testing.T) {
	session := instagramSession{}
	assert.Equal(t, baseURL, session.BaseURL())
}

func TestFetchProfileFails(t *testing.T) {
	client := resty.New()
	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()

	session := instagramSession{
		Client: client,
		Seed:   instagramProfile{},
	}

	_, err := session.FetchProfile()
	assert.NotEqual(t, nil, err)

	session = instagramSession{
		Client: client,
		Seed: instagramProfile{
			IgName: "invalid.user.name",
		},
	}
	httpmock.RegisterResponder(
		"GET",
		session.BaseURL()+"/invalid.user.name/?__a=1",
		httpmock.NewStringResponder(500, "Invalid"),
	)

	_, err = session.FetchProfile()
	assert.EqualError(t, err, "Fetch profile error")
}

func TestFetchProfileSuccess(t *testing.T) {
	client := resty.New()
	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()

	session := instagramSession{
		Client: client,
		Seed: instagramProfile{
			IgName: fakeUsername,
		},
	}

	fakeURL := fmt.Sprintf("%s/%s/?__a=1", session.BaseURL(), fakeUsername)
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
		Seed: instagramProfile{
			IgName: "invalid.user.name",
		},
	}
	httpmock.RegisterResponder(
		"GET",
		session.BaseURL()+"/graphql/query",
		httpmock.NewStringResponder(500, "Invalid"),
	)

	_, err = session.FetchRelatedProfiles(instagramProfile{})
	assert.EqualError(t, err, "Fetch related profiles error")
}

func TestFetchRelatedProfilesSuccess(t *testing.T) {
	client := resty.New()
	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()

	session := instagramSession{
		Client: client,
	}

	responder, _ := httpmock.NewJsonResponder(200, relatedProfilesFixture)
	httpmock.RegisterResponder("GET", session.BaseURL()+"/graphql/query", responder)

	profiles, err := session.FetchRelatedProfiles(instagramProfile{})
	assert.Equal(t, nil, err)

	assert.Equal(t, 4, len(profiles))
	assert.Equal(t, "1234", profiles[0].ID())
	assert.Equal(t, "2345", profiles[1].ID())
	assert.Equal(t, "3456", profiles[2].ID())
	assert.Equal(t, "4567", profiles[3].ID())
}
