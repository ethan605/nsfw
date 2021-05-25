package crawler

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSeeds(t *testing.T) {
	mockRawData := strings.NewReader("")
	assert.Panics(t, func() { parseSeeds(mockRawData) })

	mockRawData = strings.NewReader(`[{ "category": "test-category", "username": "test-username", "user_id": "123456" }]`)
	seeds := parseSeeds(mockRawData)
	assert.Equal(t, 1, len(seeds))
	assert.Equal(t, "test-category", seeds[0].Category)
	assert.Equal(t, "test-username", seeds[0].Username)
	assert.Equal(t, "123456", seeds[0].UserID)
}
