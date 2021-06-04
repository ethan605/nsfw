package crawler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProfile(t *testing.T) {
	profile := Profile{
		ID:          "1234",
		DisplayName: "Fake Name",
		Username:    "fake.user.name",
	}

	assert.Equal(t, "<Profile 1234 fake.user.name Fake Name>", profile.String())

	profile.Source = "Source"
	assert.Equal(t, "<Source 1234 fake.user.name Fake Name>", profile.String())
}
