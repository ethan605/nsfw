package crawler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestNewInstagramCrawler(t *testing.T) {
	scheduler := NewScheduler(SchedulerConfig{})

	_, err := NewInstagramCrawler(Config{}, scheduler)
	assert.EqualError(t, err, "missing required Seed config")

	_, err = NewInstagramCrawler(Config{Seed: fakeProfile}, scheduler)
	assert.EqualError(t, err, "missing required Writer config")

	config := Config{
		Seed:   fakeProfile,
		Writer: &mockWriter{},
	}
	_, err = NewInstagramCrawler(config, scheduler)
	assert.Equal(t, nil, err)
}

func TestInstagramCrawlerRunFailure(t *testing.T) {
	client := &http.Client{}
	httpmock.ActivateNonDefault(client)
	defer httpmock.DeactivateAndReset()

	seedProfile := Profile{Username: fakeUsername}
	config := Config{
		Client: client,
		Writer: &mockWriter{},
		Seed:   seedProfile,
	}
	scheduler := NewScheduler(SchedulerConfig{MaxProfiles: 3})
	crawler, _ := NewInstagramCrawler(config, scheduler)

	// FetchProfile error
	err := crawler.Run()
	assert.NotEqual(t, nil, err)

	profileResponder, _ := httpmock.NewJsonResponder(200, object{
		"graphql": object{
			"user": object{
				"id": "-1",
			},
		},
	})
	httpmock.RegisterResponder(
		"GET",
		fmt.Sprintf("/%s/?__a=1", seedProfile.Username),
		profileResponder,
	)

	// Config.Writer error on seedProfile
	err = crawler.Run()
	assert.EqualError(t, err, "error writing to output stream")

	profileResponder, _ = httpmock.NewJsonResponder(200, profileFixture)
	httpmock.RegisterResponder(
		"GET",
		fmt.Sprintf("/%s/?__a=1", seedProfile.Username),
		profileResponder,
	)

	relatedProfilesResponder, _ := httpmock.NewJsonResponder(200, generateRelatedProfilesFixture("2345", "-1"))
	httpmock.RegisterResponder(
		"GET",
		"/graphql/query",
		relatedProfilesResponder,
	)

	// Config.Writer error on related profiles
	err = crawler.Run()
	assert.EqualError(t, err, "error writing to output stream")
}

func TestInstagramCrawlerRunSuccess(t *testing.T) {
	// Setup HTTP requests mock
	client := &http.Client{}
	httpmock.ActivateNonDefault(client)
	defer httpmock.DeactivateAndReset()

	seed := Profile{Username: fakeUsername}

	profileResponder, _ := httpmock.NewJsonResponder(200, profileFixture)
	httpmock.RegisterResponder(
		"GET",
		fmt.Sprintf("/%s/?__a=1", seed.Username),
		profileResponder,
	)

	httpmock.RegisterResponder(
		"GET",
		"/graphql/query",
		func(req *http.Request) (*http.Response, error) {
			variables := object{}
			_ = json.Unmarshal([]byte(req.URL.Query().Get("variables")), &variables)
			userID, _ := variables["user_id"].(string)

			fixturesMap := object{
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

	// Setup output mock
	output := &mockWriter{}

	config := Config{
		Client: client,
		Writer: output,
		Seed:   seed,
	}
	scheduler := NewScheduler(SchedulerConfig{MaxProfiles: 4})
	crawler, _ := NewInstagramCrawler(config, scheduler)

	err := crawler.Run()
	assert.Equal(t, nil, err)

	assert.Equal(t, 5, len(output.WrittenProfiles))
	assert.Equal(t, "1234", output.WrittenProfiles[0].ID)
	assert.Equal(t, "2345", output.WrittenProfiles[1].ID)
	assert.Equal(t, "3456", output.WrittenProfiles[2].ID)
	assert.Equal(t, "4567", output.WrittenProfiles[3].ID)
	assert.Equal(t, "5678", output.WrittenProfiles[4].ID)
}

func TestFetchProfileFailure(t *testing.T) {
	client := &http.Client{}
	httpmock.ActivateNonDefault(client)
	defer httpmock.DeactivateAndReset()

	session := instagramSession{
		client: resty.NewWithClient(client),
		config: Config{
			Seed: Profile{},
		},
	}

	_, err := session.fetchProfile()
	assert.NotEqual(t, nil, err)

	session = instagramSession{
		client: resty.NewWithClient(client),
		config: Config{
			Seed: Profile{
				Username: "invalid.user.name",
			},
		},
	}
	httpmock.RegisterResponder(
		"GET",
		"/invalid.user.name/?__a=1",
		httpmock.NewStringResponder(500, "Invalid"),
	)

	_, err = session.fetchProfile()
	assert.EqualError(t, err, "fetch profile error")
}

func TestFetchRelatedProfilesFailure(t *testing.T) {
	client := &http.Client{}
	httpmock.ActivateNonDefault(client)
	defer httpmock.DeactivateAndReset()

	session := instagramSession{
		client: resty.NewWithClient(client),
	}

	_, err := session.fetchRelatedProfiles(Profile{})
	assert.NotEqual(t, nil, err)

	session = instagramSession{
		client: resty.NewWithClient(client),
		config: Config{
			Seed: Profile{
				Username: "invalid.user.name",
			},
		},
	}
	httpmock.RegisterResponder(
		"GET",
		"/graphql/query",
		httpmock.NewStringResponder(500, "Invalid"),
	)

	_, err = session.fetchRelatedProfiles(Profile{})
	assert.EqualError(t, err, "fetch related profiles error")
}

func TestInstagramSessions(t *testing.T) {
	session := instagramSession{}
	assert.Equal(t, "https://www.instagram.com", session.baseURL())
}

/* Private stuffs */

type object map[string]interface{}

var (
	fakeUsername   = "fake.user.name"
	fakeID         = "1234"
	fakeProfile    = Profile{ID: fakeID, Username: fakeUsername}
	profileFixture = object{
		"graphql": object{
			"user": object{
				"full_name":          "Fake Name",
				"id":                 fakeID,
				"profile_pic_url_hd": "https://profile-pic-url",
				"username":           fakeUsername,
			},
		},
	}
)

type mockWriter struct {
	WrittenProfiles []Profile
}

func (m *mockWriter) Write(profile Profile) error {
	if profile.ID == "-1" {
		return errors.New("error writing to output stream")
	}

	m.WrittenProfiles = append(m.WrittenProfiles, profile)
	return nil
}

func generateRelatedProfilesFixture(relatedIds ...string) object {
	edges := []object{}

	for _, id := range relatedIds {
		edges = append(edges, object{"node": object{"id": id}})
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
