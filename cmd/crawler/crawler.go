package main

import (
	"fmt"
	"nsfw/internal/crawler"

	"github.com/go-resty/resty/v2"
)

func main() {
	fmt.Println("Start crawling...")

	// TODO: read from somewhere else
	seedProfile := "vox.ngoc.traan"

	crawler.NewInstagramCrawler(
		resty.New(),
		crawler.Config{
			Defer: 2,
			Seed: crawler.NewInstagramProfile(map[string]interface{}{
				"username": seedProfile,
			}),
		},
	).Start()
}
