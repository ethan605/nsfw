package main

import (
	"log"
	"math/rand"
	"nsfw/internal/crawler"
	"runtime"
	"sync"
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
	queue := make(chan int)
	limiter := make(chan struct{}, 5)

	/* createStopSignal := func() <-chan struct{} {
		done := make(chan struct{})

		go func() {
			<-time.After(3 * time.Second)
			done <- struct{}{}
		}()

		return done
	}

	timer := createStopSignal() */

	fetchProfile := func() int {
		time.Sleep(300 * time.Millisecond)
		return 1
	}

	fetchRelatedProfiles := func(fromProfile int) []int {
		rand.Seed(time.Now().UnixNano())

		time.Sleep((time.Duration)(rand.Intn(100)+500) * time.Millisecond)
		profiles := []int{}
		numRelatedProfiles := rand.Intn(4) + 8

		for idx := 1; idx <= numRelatedProfiles; idx++ {
			profiles = append(profiles, fromProfile*100+idx)
		}

		return profiles
	}

	go func() {
		seedProfile := fetchProfile()
		log.Println(" - enqueue", seedProfile)
		queue <- seedProfile

		for _, profile := range fetchRelatedProfiles(seedProfile) {
			log.Println(" - enqueue", profile)
			queue <- profile
		}

		close(queue)
	}()

	var wg sync.WaitGroup

	for num := range queue {
		wg.Add(1)

		go func(num int) {
			limiter <- struct{}{}
			log.Println("  - processing:", num, len(limiter))
			time.Sleep(time.Second)
			<-limiter
			log.Println("    - done:", num, len(limiter))
			wg.Done()
		}(num)
	}

	wg.Wait()
	log.Println("Exitting. Goroutines:", runtime.NumGoroutine())
}
