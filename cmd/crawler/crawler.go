package main

import (
	"encoding/csv"
	"nsfw/internal/crawler"
	"os"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)

	if os.Getenv("ENV") == "production" {
		logrus.SetLevel(logrus.InfoLevel)
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

type crawlerWriter struct {
	file   *os.File
	writer *csv.Writer
}

func newCrawlerWriter() *crawlerWriter {
	file, err := os.Create("results.csv")
	panicOnError(err)

	writer := csv.NewWriter(file)

	return &crawlerWriter{
		file:   file,
		writer: writer,
	}
}

func (w *crawlerWriter) Write(profile crawler.Profile) error {
	logrus.WithField("profile", profile).Info("  - writing")

	row := []string{
		profile.ID,
		profile.Username,
		profile.DisplayName,
		profile.AvatarURL,
	}

	row = append(row, profile.Gallery...)
	return w.writer.Write(row)
}

func (w *crawlerWriter) Flush() error {
	w.writer.Flush()
	err := w.file.Close()

	if err != nil {
		logrus.WithField("error", err).Error("flushing writer failed")
	}

	return err
}

func crawlInstagram(mock bool) {
	writer := newCrawlerWriter()
	defer writer.Flush()

	// TODO: read from somewhere else
	seedInstagramProfile := crawler.Profile{
		ID:       "3030197091",
		Username: "vox.ngoc.traan",
	}

	config := crawler.Config{
		Seed:   seedInstagramProfile,
		Writer: writer,
	}

	schedulerConfig := crawler.SchedulerConfig{
		DeferTime:   time.Second,
		MaxProfiles: 20,
		MaxWorkers:  1,
	}
	scheduler := crawler.NewScheduler(schedulerConfig)
	instagramCrawler, err := crawler.NewInstagramCrawler(config, scheduler)

	if mock {
		limiterConfig := crawler.LimiterConfig{
			DeferTime:  time.Second,
			MaxTakes:   10,
			MaxWorkers: 3,
		}
		limiter := crawler.NewLimiter(limiterConfig)
		instagramCrawler, err = mockInstagramCrawler(config, limiter)
	}

	panicOnError(err)

	instagramCrawler.Run()
}

func panicOnError(err error) {
	if err != nil {
		logrus.Panicln(err)
	}
}
