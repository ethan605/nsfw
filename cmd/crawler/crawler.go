package main

import (
	"fmt"
	"log"
	"nsfw/internal/crawler"
	"os"
)

func main() {
	fmt.Println("Start crawling...")

	seedsSource, err := os.Open("./configs/crawler/seeds/instagram.json")

	if err != nil {
		log.Panicln("Error reading seed source")
	}

	source := crawler.Source{Name: "instagram", RawData: seedsSource}
	_ = crawler.Crawl(source)
}
