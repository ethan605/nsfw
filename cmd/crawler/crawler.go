package main

import (
	"nsfw/internal/crawler"
	"os"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)

	if os.Getenv("ENV") == "production" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}
}

func main() {
	defer func() {
		logrus.
			WithFields(logrus.Fields{"goroutines": runtime.NumGoroutine()}).
			Info("Gracefully shutting down")
	}()

	crawlInstagram(true)
}

/* Private stuffs */

type crawlerWriter struct{}

func (o *crawlerWriter) Write(profile crawler.Profile) error {
	logrus.
		WithFields(logrus.Fields{"profile": profile}).
		Info("  writing")
	return nil
}

func crawlInstagram(mock bool) {
	// TODO: read from somewhere else
	seedInstagramUsername := "vox.ngoc.traan"

	config := crawler.Config{
		Writer: &crawlerWriter{},
		Seed:   crawler.NewInstagramSeed(seedInstagramUsername),
	}

	scheduler := crawler.NewScheduler(time.Second, 100)
	instagramCrawler, err := crawler.NewInstagramCrawler(config, scheduler)

	if mock {
		scheduler := crawler.NewScheduler(time.Second/2, 10)
		instagramCrawler, err = mockInstagramCrawler(config, scheduler)
	}

	panicOnError(err)

	err = instagramCrawler.Run()
	panicOnError(err)
}

func panicOnError(err error) {
	if err != nil {
		logrus.Panicln(err)
	}
}
