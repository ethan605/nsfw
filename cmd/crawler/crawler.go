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
	// testConcurrency()
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

	schedulerConfig := crawler.SchedulerConfig{
		DeferTime:   time.Second,
		MaxProfiles: 100,
		MaxWorkers:  3,
	}
	scheduler := crawler.NewScheduler(schedulerConfig)
	instagramCrawler, err := crawler.NewInstagramCrawler(config, scheduler)

	if mock {
		schedulerConfig := crawler.SchedulerConfig{
			DeferTime:   time.Second / 2,
			MaxProfiles: 10,
			MaxWorkers:  1,
		}
		scheduler := crawler.NewScheduler(schedulerConfig)
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

/* func testConcurrency() {
	maxJobs := 5
	queue := make(chan int, maxJobs)
	limiter := make(chan struct{}, maxJobs)

	producer := func() {
		idx := 0

		for {
			time.Sleep(100 * time.Millisecond)
			queue <- idx
			idx++

			if idx >= 20 {
				close(queue)
				return
			}
		}
	}

	go producer()

	wg := &sync.WaitGroup{}

	for num := range queue {
		limiter <- struct{}{}

		wg.Add(1)
		go func(num int) {
			defer func() {
				<-limiter
				wg.Done()
			}()

			time.Sleep(time.Second)
			logrus.
				WithFields(logrus.Fields{"num": num}).
				Info("queue")
		}(num)
	}

	wg.Wait()
} */
