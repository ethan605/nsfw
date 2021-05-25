package main

import (
	"fmt"
	"nsfw/internal/crawler"

	"github.com/go-resty/resty/v2"
)

func main() {
	fmt.Println("Start crawling...")
	instagramCrawler := crawler.NewInstagramCrawler(resty.New(), crawler.SeedProfile)
	instagramCrawler.Crawl()
}
