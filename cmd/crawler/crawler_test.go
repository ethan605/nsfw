package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCrawler(t *testing.T) {
	assert.Panics(t, func() { main() }, "open ./configs/crawler/seeds/instagram.json: no such file or directory")
}
