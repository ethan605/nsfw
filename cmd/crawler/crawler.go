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
	crawlInstagram(true)
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
	// TODO: read from somewhere else
	seedInstagramUsername := "vox.ngoc.traan"

	config := crawler.Config{
		Output: &crawlerOutput{},
		Seed:   crawler.NewInstagramSeed(seedInstagramUsername),
	}
	scheduler := crawler.NewScheduler(time.Second, 10)

	instagramCrawler, err := crawler.NewInstagramCrawler(config, scheduler)

	if dryRun {
		instagramCrawler, err = mockInstagramCrawler(config, scheduler)
	}

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
