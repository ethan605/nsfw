package crawler

import (
	"fmt"
	"testing"
	"testing/fstest"
)

func TestInstagramSeeds(t *testing.T) {
	fs := fstest.MapFS{
		"./instagram.json": {
			Data: []byte("hello, world"),
		},
	}

	b, _ := fs.ReadFile("./instagram.json")
	fmt.Printf("m.ReadFile: %+v\n", string(b))

	// seeds := parseInstagramSeeds(m)
	// assert.Equal(t, 3, len(seeds))
}
