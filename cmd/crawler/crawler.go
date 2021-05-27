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

func expGoroutines() {
	queue := make(chan int, 5)
	// done := make(chan struct{})
	timer := time.After(3 * time.Second)

	go func() {
		num := 0

		for {
			select {
			case queue <- num:
				log.Println("enqueue:", num)
				time.Sleep(100 * time.Millisecond)
				num++
			case <-timer:
				log.Println("Break signal received")
				close(queue)
				return
			}
		}
	}()

	for num := range queue {
		go func(num int) {
			log.Println("  - processing:", num, len(queue))
			time.Sleep(2 * time.Second)
			log.Println("  - done:", num, len(queue))
		}(num)
		time.Sleep(time.Second)
	}

	log.Println("Exitting. Goroutines:", runtime.NumGoroutine())
}
