package crawler

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInstagram(t *testing.T) {
	mockSource := strings.NewReader("[{ \"category\": \"test\", \"username\": \"test\" }]")
	assert.NotPanics(t, func() { Instagram(mockSource) })
}
