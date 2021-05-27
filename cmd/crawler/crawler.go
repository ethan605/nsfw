package main

import (
	"log"
	"nsfw/internal/crawler"
	"runtime"
	"time"
)

func main() {
	crawlInstagram(true)
	expGoroutines()
}

/* Private stuffs */

type crawlerOutput struct{}

func (o *crawlerOutput) Write(profile crawler.Profile) error {
	log.Println(" - Crawled profile:", profile)
	return nil
}

func crawlInstagram(dryRun bool) {
	if dryRun {
		return
	}

	// TODO: read from somewhere else
	seedProfile := "vox.ngoc.traan"

	config := crawler.Config{
		Defer:  2,
		Output: &crawlerOutput{},
		Seed: crawler.NewInstagramProfile(map[string]interface{}{
			"Username": seedProfile,
		}),
	}

	instagramCrawler, err := crawler.NewInstagramCrawler(config)
	panicOnError(err)

	err = instagramCrawler.Start()
	panicOnError(err)
}

func panicOnError(err error) {
	if err != nil {
		log.Panicln(err)
	}
}

func createStopSignal() <-chan struct{} {
	done := make(chan struct{})
	// sig := make(chan os.Signal, 1)
	// signal.Notify(sig, os.Interrupt)

	go func() {
		<-time.After(3 * time.Second)
		// <-sig
		done <- struct{}{}
	}()

	return done
}

func expGoroutines() {
	queue := make(chan int)
	limiter := make(chan struct{}, 5)
	stopper := createStopSignal()
	done := make(chan struct{})

	producer := func(quit <-chan struct{}) {
		num := 0

		for {
			select {
			case <-quit:
				log.Println("[Stop signal received]")
				close(queue)
				return
			default:
				num++
				log.Println("enqueue:", num)
				time.Sleep(100 * time.Millisecond)
				queue <- num
			}
		}
	}

	go producer(stopper)

	for num := range queue {
		go func(num int) {
			limiter <- struct{}{}
			log.Println("  - processing:", num, len(limiter))
			time.Sleep(time.Second)
			<-limiter
			log.Println("    - done:", num, len(limiter))

			if len(limiter) == 0 {
				close(done)
			}
		}(num)
	}

	<-done
	log.Println("Exitting. Goroutines:", runtime.NumGoroutine())
}
