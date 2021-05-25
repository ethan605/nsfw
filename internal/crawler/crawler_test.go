package crawler

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCrawl(t *testing.T) {
	err := Crawl(Source{Name: "invalid-name"})
	assert.Error(t, err, "Invalid")

	mockRawData := strings.NewReader("[{}]")
	err = Crawl(Source{Name: "instagram", RawData: mockRawData})
	assert.Nil(t, err)
}

func TestParseSeeds(t *testing.T) {
	mockRawData := strings.NewReader("{}")
	assert.Panics(t, func() { parseSeeds(mockRawData) }, "panic: runtime error: invalid memory address or nil pointer dereference")

	mockRawData = strings.NewReader(`[{ "category": "test-category", "username": "test-username", "user_id": "123456" }]`)
	seeds := parseSeeds(mockRawData)
	assert.Equal(t, 1, len(seeds))
	assert.Equal(t, "test-category", seeds[0].Category)
	assert.Equal(t, "test-username", seeds[0].Username)
	assert.Equal(t, "123456", seeds[0].UserID)
}
