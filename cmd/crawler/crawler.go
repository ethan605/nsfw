package main

import (
	"fmt"
	"log"
	"nsfw/internal/crawler"
	"os"

	"github.com/go-resty/resty/v2"
)

func main() {
	fmt.Println("Start crawling...")

	source, err := os.Open("./configs/crawler/seeds/instagram.json")
	if err != nil {
		log.Panicln("Error reading seed source")
	}

	instagramCrawler := crawler.NewInstagramCrawler(resty.New(), source)
	instagramCrawler.Crawl()
}
