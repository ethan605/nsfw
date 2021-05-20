package crawler

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
)

func TestInstagramSeeds(t *testing.T) {
	fs := fstest.MapFS{
		"./instagram.json": {
			Data: []byte("hello, world"),
		},
	}

	b, _ := fs.ReadFile("./instagram.json")
	// fmt.Printf("m.ReadFile: %+v\n", string(b))
	assert.Equal(t, nil, b)

	// seeds := parseInstagramSeeds(m)
	// assert.Equal(t, 3, len(seeds))
}
