package crawler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInstagram(t *testing.T) {
	assert.Equal(t, Instagram(), 0)
}
