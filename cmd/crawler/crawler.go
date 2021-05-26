package main

import (
	"nsfw/internal/crawler"
)

func main() {
	crawlInstagram(false)
}

func crawlInstagram(dryRun bool) {
	if dryRun {
		return
	}

	// TODO: read from somewhere else
	seedProfile := "vox.ngoc.traan"

	config := crawler.Config{
		Defer: 2,
		Seed: crawler.NewInstagramProfile(map[string]interface{}{
			"Username": seedProfile,
		}),
	}

	instagramCrawler, _ := crawler.NewInstagramCrawler(config)
	_ = instagramCrawler.Start()
}
