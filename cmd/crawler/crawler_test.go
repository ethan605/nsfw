package main

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCrawler(t *testing.T) {
	t.Skip()
	assert.NotPanics(t, func() { main() })
}

func TestPanicOnError(t *testing.T) {
	assert.NotPanics(t, func() { panicOnError(nil) })
	assert.Panics(t, func() { panicOnError(errors.New("Fake error")) })
}
