package crawler

import (
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

var (
	profileFixture = map[string]interface{}{
		"graphql": map[string]interface{}{
			"user": map[string]string{
				"full_name":          "Fake Name",
				"id":                 "1234",
				"profile_pic_url_hd": "https://profile-pic-url",
				"username":           "fake.user.name",
			},
		},
	}
	relatedProfilesFixture = map[string]interface{}{
		"data": map[string]interface{}{
			"user": map[string]interface{}{
				"edge_chaining": map[string]interface{}{
					"edges": []map[string]interface{}{
						{"node": map[string]string{"id": "1234"}},
						{"node": map[string]string{"id": "2345"}},
						{"node": map[string]string{"id": "3456"}},
						{"node": map[string]string{"id": "4567"}},
					},
				},
			},
		},
	}
)

func TestInstagramCrawler(t *testing.T) {
	client := resty.New()
	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()

	mockSource := strings.NewReader(`[{
		"category": "fake-category",
		"username": "fake.user.name"
	}]`)
	session := NewInstagramCrawler(client, mockSource)

	profileResponder, _ := httpmock.NewJsonResponder(200, profileFixture)
	httpmock.RegisterResponder(
		"GET",
		"https://www.instagram.com/fake.user.name/?__a=1",
		profileResponder,
	)

	relatedProfilesResponder, _ := httpmock.NewJsonResponder(200, relatedProfilesFixture)
	httpmock.RegisterResponder(
		"GET",
		"https://www.instagram.com/graphql/query",
		relatedProfilesResponder,
	)

	assert.NotPanics(t, func() { session.Crawl() })
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
	}

	_, err := session.FetchProfile()
	assert.NotEqual(t, nil, err)

	session = instagramSession{
		Client: client,
		Seed: crawlSeed{
			Username: "invalid.user.name",
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
		Seed: crawlSeed{
			Category: "fake-category",
			Username: "fake.user.name",
		},
	}

	fakeURL := session.BaseURL() + "/fake.user.name/?__a=1"
	responder, _ := httpmock.NewJsonResponder(200, profileFixture)
	httpmock.RegisterResponder("GET", fakeURL, responder)

	profile, err := session.FetchProfile()
	assert.Equal(t, nil, err)
	assert.Equal(t, "https://profile-pic-url", profile.AvatarURL())
	assert.Equal(t, "1234", profile.ID())
	assert.Equal(t, "Fake Name", profile.DisplayName())
	assert.Equal(t, "fake.user.name", profile.Username())
	assert.Equal(t, "fake-category", profile.Category())
	assert.Equal(t, "<Instagram fake-category 1234 fake.user.name Fake Name>", profile.String())
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
		Seed: crawlSeed{
			Username: "invalid.user.name",
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
		Seed: crawlSeed{
			Category: "fake-category",
		},
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

	for _, profile := range profiles {
		assert.Equal(t, "fake-category", profile.Category())
	}
}
