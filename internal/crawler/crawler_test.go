package crawler

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPanicOnError(t *testing.T) {
	assert.NotPanics(t, func() { panicOnError(nil) })
	assert.Panics(t, func() { panicOnError(errors.New("Fake error")) })
}
