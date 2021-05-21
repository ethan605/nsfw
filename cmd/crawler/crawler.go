package main

import (
	"fmt"
	"nsfw/internal/crawler"
	"os"
)

func main() {
	fmt.Println("Start crawling...")

	seedsSource, _ := os.Open("./configs/crawler/seeds/instagram.json")
	source := crawler.Source{Name: "instagram", RawData: seedsSource}
	_ = crawler.Crawl(source)
}
