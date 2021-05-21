package crawler

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCrawlInstagram(t *testing.T) {
	mockSource := strings.NewReader("[{ \"category\": \"test\", \"username\": \"test\", \"user_id\": \"123456\" }]")
	assert.NotPanics(t, func() { CrawlInstagram(mockSource) })
}

func TestInstagramSessions(t *testing.T) {
	session := instagramSession{}

	assert.Equal(t, "https://instagram.com", session.BaseURL())
	assert.Equal(t, "https://instagram.com/test-username/?__a=1", session.ProfileURL("test-username"))
}
