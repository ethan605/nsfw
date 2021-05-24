package crawler

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

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

func TestCrawlInstagram(t *testing.T) {
	mockSource := strings.NewReader("[{}]")
	assert.NotPanics(t, func() { crawlInstagram(mockSource) })
}

func TestInstagramSessions(t *testing.T) {
	session := instagramSession{
		seedStruct: seedStruct{},
	}

	assert.Equal(t, "https://instagram.com", session.BaseURL())

	/* client := &mockClient{response: "fake-profile-response"}
	assert.Equal(t, "fake-profile-response", session.FetchProfile(client))

	client = &mockClient{response: "fake-related-profiles-response"}
	assert.Equal(
		t,
		"fake-related-profiles-response",
		session.FetchRelatedProfiles(client, ""),
	) */
}
