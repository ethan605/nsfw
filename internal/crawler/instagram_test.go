package crawler

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

type mockClient struct {
	response string
}

func (m *mockClient) Do(req *http.Request) (*http.Response, error) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(bytes.NewBufferString(m.response)),
	}

	return resp, nil
}

/* func TestInstagramCrawler(t *testing.T) {
	mockSource := strings.NewReader("[{}]")
	assert.NotPanics(t, func() { crawlInstagram(mockSource) })
} */

func TestInstagramSessions(t *testing.T) {
	session := instagramSession{}
	assert.Equal(t, "https://www.instagram.com", session.BaseURL())
}

func TestFetchProfileFail(t *testing.T) {
	client := resty.New()
	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()

	session := instagramSession{
		RestyClient: client,
	}

	_, err := session.FetchProfile()
	assert.NotEqual(t, nil, err)

	session = instagramSession{
		RestyClient: client,
		Seed: seedStruct{
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
		RestyClient: client,
		Seed: seedStruct{
			Category: "fake-category",
			Username: "fake.user.name",
		},
	}

	fixture := map[string]interface{}{
		"graphql": map[string]interface{}{
			"user": map[string]string{
				"full_name":          "Fake Name",
				"id":                 "1234",
				"profile_pic_url_hd": "https://profile-pic-url",
				"username":           "fake.user.name",
			},
		},
	}
	fakeURL := session.BaseURL() + "/fake.user.name/?__a=1"
	responder, _ := httpmock.NewJsonResponder(200, fixture)
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

func TestFetchRelatedProfiles(t *testing.T) {
	session := instagramSession{
		Client: &mockClient{
			response: `OK`,
		},
	}

	_, err := session.FetchRelatedProfiles(instagramProfile{})
	assert.NotEqual(t, nil, err)

	session = instagramSession{
		Client: &mockClient{
			response: `{
				"data": {
					"user": {
						"edge_chaining": {
							"edges": [
								{"node": {"id": "1234"}},
								{"node": {"id": "2345"}},
								{"node": {"id": "3456"}},
								{"node": {"id": "4567"}}
							]
						}
					}
				}
			}`,
		},
		Seed: seedStruct{Category: "fake-category"},
	}

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
