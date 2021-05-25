package main

import (
	"fmt"
	"log"
	"net/http"
	"nsfw/internal/crawler"
	"os"
)

func main() {
	fmt.Println("Start crawling...")

	seedsSource, err := os.Open("./configs/crawler/seeds/instagram.json")

	if err != nil {
		log.Panicln("Error reading seed source")
	}

	instagramCrawler := crawler.InstagramCrawler{
		Source: seedsSource,
		Client: &http.Client{},
	}

	instagramCrawler.Crawl()
}
