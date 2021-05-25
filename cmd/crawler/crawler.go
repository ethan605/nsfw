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

	instagramCrawler := crawler.NewInstagramCrawler(resty.New(), seedProfile)
	instagramCrawler.Crawl()
}
