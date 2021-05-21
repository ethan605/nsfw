package crawler

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInstagramSeeds(t *testing.T) {
	mockFile := strings.NewReader("")
	assert.Panics(t, func() { parseSeeds(mockFile) }, "panic: runtime error: invalid memory address or nil pointer dereference")

	mockFile = strings.NewReader("[{ \"category\": \"test-category\", \"username\": \"test-username\" }]")
	seeds := parseSeeds(mockFile)
	assert.Equal(t, 1, len(seeds))
	assert.Equal(t, "test-category", seeds[0].Category)
	assert.Equal(t, "test-username", seeds[0].Username)
}
