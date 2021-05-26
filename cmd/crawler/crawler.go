package main

import (
	"log"
	"nsfw/internal/crawler"
)

func main() {
	crawlInstagram(false)
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
