package crawler

import (
	"errors"
	"fmt"
	"io/ioutil"
	"runtime"
	"sort"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func init() {
	logrus.SetOutput(ioutil.Discard)
}

func TestSchedulerFailure(t *testing.T) {
	initialGoRoutines := runtime.NumGoroutine()

	scheduler := newScheduler(time.Millisecond, 10)

	seedProfile := instagramProfile{UserID: "-1"}
	go scheduler.Run(mockFetchProfiles, seedProfile)

	profileIDs := []string{}

	for profile := range scheduler.Results() {
		profileIDs = append(profileIDs, profile.ID())
	}

	assert.Equal(t, 0, len(profileIDs))
	assert.Equal(t, initialGoRoutines, runtime.NumGoroutine())
}

func TestSchedulerSuccess(t *testing.T) {
	initialGoRoutines := runtime.NumGoroutine()

	scheduler := newScheduler(time.Millisecond, 10)
	assert.NotEqual(t, nil, scheduler)

	seedProfile := instagramProfile{UserID: "1"}
	go scheduler.Run(mockFetchProfiles, seedProfile)

	profileIDs := []string{}

	for profile := range scheduler.Results() {
		profileIDs = append(profileIDs, profile.ID())
	}

	assert.Equal(t, 12, len(profileIDs))

	// First level
	results := []string{"1/1", "1/2", "1/3"}
	assert.Equal(t, results, profileIDs[0:3])

	// Second level
	results = []string{
		"1/1/1", "1/1/2", "1/1/3",
		"1/2/1", "1/2/2", "1/2/3",
		"1/3/1", "1/3/2", "1/3/3",
	}
	sort.Strings(profileIDs[3:])
	assert.Equal(t, results, profileIDs[3:])

	assert.Equal(t, initialGoRoutines, runtime.NumGoroutine())
}

/* Private stuffs */

func mockFetchProfiles(fromProfile Profile) ([]Profile, error) {
	if fromProfile.ID() == "-1" {
		return nil, errors.New("fake error")
	}

	profiles := []Profile{}

	for idx := 1; idx <= 3; idx++ {
		relatedProfile := instagramProfile{
			UserID: fmt.Sprintf("%s/%d", fromProfile.ID(), idx),
		}
		profiles = append(profiles, relatedProfile)
	}

	return profiles, nil
}
