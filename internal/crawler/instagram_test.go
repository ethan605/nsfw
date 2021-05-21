package crawler

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCrawlInstagram(t *testing.T) {
	mockSource := strings.NewReader("[]")
	assert.NotPanics(t, func() { crawlInstagram(mockSource) })
}

func TestInstagramSessions(t *testing.T) {
	session := instagramSession{sessionID: "fake-session", suggestedQueryHash: "fake-query-hash"}

	assert.Equal(t, "https://instagram.com", session.BaseURL())
	assert.Equal(t, "https://instagram.com/fake-username/?__a=1", session.FetchProfile("fake-username"))
	assert.Equal(t, "https://instagram.com/?user_id=fake-username&session=fake-session&query_hash=fake-query-hash", session.FetchOtherUsers("fake-username"))
}
