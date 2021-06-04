package crawler

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDummyCrawler(t *testing.T) {
	_, err := NewDummyCrawler(Config{}, LimiterConfig{})
	assert.EqualError(t, err, "missing required Seed config")

	seedProfile := Profile{ID: "1"}

	_, err = NewDummyCrawler(Config{Seed: seedProfile}, LimiterConfig{})
	assert.EqualError(t, err, "missing required Writer config")

	config := Config{
		Seed:   seedProfile,
		Writer: &mockWriter{},
	}
	_, err = NewDummyCrawler(config, LimiterConfig{})
	assert.Equal(t, nil, err)
}

func TestDummyCrawlerFailure(t *testing.T) {
	config := Config{
		Seed:   Profile{ID: "-1"},
		Writer: &mockWriter{},
	}
	limiterConfig := LimiterConfig{
		MaxTakes: 3,
	}

	crawler, _ := NewDummyCrawler(config, limiterConfig)
	assert.NotPanics(t, crawler.Run)
}

func TestDummyCrawlerSuccess(t *testing.T) {
	config := Config{
		Seed:   Profile{ID: "1"},
		Writer: &mockWriter{},
	}
	limiterConfig := LimiterConfig{
		MaxTakes: 3,
	}

	crawler, _ := NewDummyCrawler(config, limiterConfig)
	assert.NotPanics(t, crawler.Run)
}

/* Private stuffs */

type mockWriter struct {
	WrittenProfiles []Profile
}

func (m *mockWriter) Write(profile Profile) error {
	if profile.ID == "-1" {
		return errors.New("error writing to output stream")
	}

	m.WrittenProfiles = append(m.WrittenProfiles, profile)
	return nil
}
