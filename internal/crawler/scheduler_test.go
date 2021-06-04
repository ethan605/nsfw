package crawler

import (
	"errors"
	"fmt"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSchedulerFailure(t *testing.T) {
	initialGoRoutines := runtime.NumGoroutine()
	scheduler := NewScheduler(SchedulerConfig{MaxProfiles: 4})

	seedProfile := Profile{ID: "-1"}
	go scheduler.Run(mockCrawl, seedProfile)

	profileIDs := []string{}

	for profile := range scheduler.Results() {
		profileIDs = append(profileIDs, profile.ID)
	}

	assert.Equal(t, 1, len(profileIDs))
	assert.Equal(t, initialGoRoutines, runtime.NumGoroutine())
}

func TestSchedulerSuccess(t *testing.T) {
	initialGoRoutines := runtime.NumGoroutine()

	scheduler := NewScheduler(SchedulerConfig{MaxProfiles: 4})
	assert.NotEqual(t, nil, scheduler)

	seedProfile := Profile{ID: "1"}
	go scheduler.Run(mockCrawl, seedProfile)

	profileIDs := []string{}

	for profile := range scheduler.Results() {
		profileIDs = append(profileIDs, profile.ID)
	}

	assert.Equal(t, 4, len(profileIDs))

	results := []string{
		"1",
		"1/1",
		"1/2",
		"1/3",
	}
	sort.Strings(profileIDs)
	assert.Equal(t, results, profileIDs)

	assert.Equal(t, initialGoRoutines, runtime.NumGoroutine())
}

/* Private stuffs */

func mockCrawl(fromProfile Profile) (Profile, []Profile, error) {
	if strings.HasPrefix(fromProfile.ID, "-1/") {
		return Profile{}, nil, errors.New("fake error")
	}

	profiles := []Profile{}

	for idx := 1; idx <= 3; idx++ {
		relatedProfile := Profile{
			ID: fmt.Sprintf("%s/%d", fromProfile.ID, idx),
		}
		profiles = append(profiles, relatedProfile)
	}

	return fromProfile, profiles, nil
}
