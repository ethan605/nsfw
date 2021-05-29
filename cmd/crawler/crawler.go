package main

import (
	"nsfw/internal/crawler"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

func main() {
	crawlInstagram(false)
}

/* Private stuffs */

type crawlerOutput struct{}

func (o *crawlerOutput) Write(profile crawler.Profile) error {
	logrus.
		WithFields(logrus.Fields{"profile": profile}).
		Debug(" writing")
	return nil
}

func crawlInstagram(dryRun bool) {
	if dryRun {
		return
	}

	// TODO: read from somewhere else
	seedInstagramUsername := "vox.ngoc.traan"

	config := crawler.Config{
		DeferTime:   time.Second / 2,
		MaxProfiles: 10,
		Output:      &crawlerOutput{},
		Seed:        crawler.NewInstagramSeed(seedInstagramUsername),
	}

	instagramCrawler, err := crawler.MockInstagramCrawler(config)
	panicOnError(err)

	err = instagramCrawler.Start()
	panicOnError(err)

	logrus.
		WithFields(logrus.Fields{"goroutines": runtime.NumGoroutine()}).
		Debug("Exitting")
}

func panicOnError(err error) {
	if err != nil {
		logrus.Panicln(err)
	}
}
